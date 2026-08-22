package tests

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/db"
	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/category"
	"eka-dev.cloud/master-data/modules/internalModule"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/promotion"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/common"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	minioClient "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	sharedDB       *sqlx.DB
	sharedTeardown func()
	setupOnce      sync.Once
	setupErr       error

	sharedMinioClient   *minioClient.Client
	sharedMinioTeardown func()
	setupMinioOnce      sync.Once
	setupMinioErr       error
)

func SetupTestPostgres(t *testing.T) (*sqlx.DB, func()) {
	setupOnce.Do(func() {
		ctx := context.Background()
		postgresContainer, err := postgres.Run(ctx,
			"postgres:15-alpine",
			postgres.WithDatabase("master_data_test"),
			postgres.WithUsername("postgres"),
			postgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second)),
		)
		if err != nil {
			setupErr = fmt.Errorf("Docker/Testcontainers unavailable: %v", err)
			return
		}

		connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			setupErr = fmt.Errorf("failed to get connection string: %v", err)
			return
		}

		dbConn, err := sqlx.Connect("postgres", connStr)
		if err != nil {
			setupErr = fmt.Errorf("failed to connect to test postgres database: %v", err)
			return
		}

		// Apply all .up.sql database schema migrations
		migrationFiles, _ := filepath.Glob("../db/migrations/*.up.sql")
		if len(migrationFiles) == 0 {
			migrationFiles, _ = filepath.Glob("db/migrations/*.up.sql")
		}
		sort.Strings(migrationFiles)
		for _, f := range migrationFiles {
			sqlContent, err := os.ReadFile(f)
			if err == nil && len(sqlContent) > 0 {
				_, _ = dbConn.Exec(string(sqlContent))
			}
		}

		// Reset serial sequences to prevent ID clashes with hardcoded seed IDs
		_, _ = dbConn.Exec(`
			SELECT setval(pg_get_serial_sequence('tm_categories', 'id'), COALESCE((SELECT MAX(id) FROM tm_categories), 1) + 1, false);
			SELECT setval(pg_get_serial_sequence('tm_menus', 'id'), COALESCE((SELECT MAX(id) FROM tm_menus), 1) + 1, false);
			SELECT setval(pg_get_serial_sequence('tm_tables', 'id'), COALESCE((SELECT MAX(id) FROM tm_tables), 1) + 1, false);
			SELECT setval(pg_get_serial_sequence('tm_promotions', 'id'), COALESCE((SELECT MAX(id) FROM tm_promotions), 1) + 1, false);
		`)

		// Set package db.DB global pointer to real test DB
		db.DB = dbConn
		sharedDB = dbConn

		sharedTeardown = func() {
			_ = dbConn.Close()
			_ = postgresContainer.Terminate(ctx)
		}
	})

	if setupErr != nil {
		t.Skipf("Skipping integration test: %v", setupErr)
	}

	return sharedDB, func() {}
}

func SetupTestMinio(t *testing.T) (*minioClient.Client, func()) {
	setupMinioOnce.Do(func() {
		ctx := context.Background()
		minioContainer, err := tcminio.Run(ctx,
			"minio/minio:RELEASE.2024-01-16T16-07-38Z",
			tcminio.WithUsername("minioadmin"),
			tcminio.WithPassword("minioadmin"),
		)
		if err != nil {
			setupMinioErr = fmt.Errorf("failed to start MinIO testcontainer: %v", err)
			return
		}

		endpoint, err := minioContainer.ConnectionString(ctx)
		if err != nil {
			setupMinioErr = fmt.Errorf("failed to get MinIO connection string: %v", err)
			return
		}

		mClient, err := minioClient.New(endpoint, &minioClient.Options{
			Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
			Secure: false,
		})
		if err != nil {
			setupMinioErr = fmt.Errorf("failed to create MinIO client: %v", err)
			return
		}

		// Ensure bucket 'coffe' exists
		bucketName := "coffe"
		exists, err := mClient.BucketExists(ctx, bucketName)
		if err == nil && !exists {
			_ = mClient.MakeBucket(ctx, bucketName, minioClient.MakeBucketOptions{})
		}

		config.Config.MinioEndpoint = endpoint
		config.Config.MinioAccessKey = "minioadmin"
		config.Config.MinioSecretKey = "minioadmin"
		config.Config.MinioBucketName = bucketName
		config.Config.MinioBaseURL = "https://storage.eka-dev.cloud/project"

		lib.SetMinioClient(mClient)
		sharedMinioClient = mClient

		sharedMinioTeardown = func() {
			_ = minioContainer.Terminate(ctx)
		}
	})

	if setupMinioErr != nil {
		t.Skipf("Skipping MinIO integration test: %v", setupMinioErr)
	}

	return sharedMinioClient, func() {}
}

func SetupTestApp(dbConn *sqlx.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(response.Success("OK", nil))
	})

	// Real Module Handlers (No Mocks!)
	category.NewHandler(app, dbConn)
	menu.NewHandler(app, dbConn)
	upload.NewHandler(app)
	table.NewHandler(app, dbConn)
	internalModule.NewHandler(app, dbConn)
	promotion.NewHandler(app, dbConn)

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	return app
}

func GenerateTestToken(userId int64, email, role string) string {
	secret := config.Config.SecretJwt
	if secret == "" {
		secret = "super-secret-jwt-key"
		config.Config.SecretJwt = secret
	}
	claims := common.Claims{
		FullName: "Test User",
		Email:    email,
		UserId:   userId,
		Type:     "ACCESS",
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func GenerateHMACSignature(queryString, bodyString, timestamp string) string {
	secret := config.Config.Secret
	if secret == "" {
		secret = "super-secret-key"
		config.Config.Secret = secret
	}
	message := fmt.Sprintf("%s%s%s", queryString, timestamp, bodyString)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ExecuteTestRequest(app *fiber.App, method, url string, body []byte, token string) (*http.Response, error) {
	var req *http.Request
	if len(body) > 0 {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return app.Test(req)
}
