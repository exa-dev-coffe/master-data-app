package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"eka-dev.cloud/master-data/db"
	"eka-dev.cloud/master-data/middleware"
	"eka-dev.cloud/master-data/modules/category"
	"eka-dev.cloud/master-data/modules/internalModule"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/table"
	"eka-dev.cloud/master-data/modules/upload"
	"eka-dev.cloud/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestPostgres(t *testing.T) (*sqlx.DB, func()) {
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
		t.Skipf("Skipping integration test: Docker/Testcontainers unavailable: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	dbConn, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test postgres database: %v", err)
	}

	// Apply all .up.sql database schema migrations
	migrationFiles, _ := filepath.Glob("db/migrations/*.up.sql")
	sort.Strings(migrationFiles)
	for _, f := range migrationFiles {
		sqlContent, err := os.ReadFile(f)
		if err == nil && len(sqlContent) > 0 {
			_, _ = dbConn.Exec(string(sqlContent))
		}
	}

	// Set package db.DB global pointer to real test DB
	db.DB = dbConn

	teardown := func() {
		_ = dbConn.Close()
		_ = postgresContainer.Terminate(ctx)
	}

	return dbConn, teardown
}

func setupTestApp(dbConn *sqlx.DB) *fiber.App {
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

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	return app
}

func executeTestRequest(app *fiber.App, method, url string, body []byte) (*http.Response, error) {
	var req *http.Request
	if len(body) > 0 {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	return app.Test(req)
}
