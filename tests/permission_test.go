package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"eka-dev.cloud/master-data/config"
	"eka-dev.cloud/master-data/lib"
	"eka-dev.cloud/master-data/modules/category"
	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/utils/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func generateCustomToken(userId int64, role string, roleId int, permissions map[string]common.PermissionAction) string {
	secret := config.Config.SecretJwt
	if secret == "" {
		secret = "super-secret-jwt-key"
		config.Config.SecretJwt = secret
	}

	if lib.RedisClient != nil && permissions != nil {
		permBytes, _ := json.Marshal(permissions)
		lib.RedisClient.Set(context.Background(), fmt.Sprintf("auth:role_permissions:%d", roleId), string(permBytes), 24*time.Hour)
	}

	claims := common.Claims{
		FullName: "Custom Perm User",
		Email:    "custom@perm.test",
		UserId:   userId,
		Type:     "ACCESS",
		Role:     role,
		RoleId:   roleId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestPBACPermissionSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)

	t.Run("Super Admin bypasses permission check", func(t *testing.T) {
		token := GenerateTestToken(1, "admin@coffe.com", "admin")

		reqBody := category.CreateCategoryRequest{
			Name: "Admin Exclusive Category",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", bodyBytes, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Barista denied when catalog:create is false", func(t *testing.T) {
		perms := map[string]common.PermissionAction{
			"catalog": {View: true, Create: false, Edit: false, Delete: false},
		}
		token := generateCustomToken(3, "barista_nocreate@coffe.com", 3, perms)

		isAvail := true
		reqBody := menu.CreateMenuRequest{
			Name:        "Unauthorized Coffee",
			Price:       35000,
			Description: "Should be denied",
			Photo:       "http://example.com/coffee.jpg",
			IsAvailable: &isAvail,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", bodyBytes, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Barista allowed when catalog:create is granted", func(t *testing.T) {
		perms := map[string]common.PermissionAction{
			"catalog": {View: true, Create: true, Edit: true, Delete: false},
		}
		token := generateCustomToken(3, "barista_withcreate@coffe.com", 3, perms)

		isAvail := true
		reqBody := menu.CreateMenuRequest{
			Name:        "Barista Special Blend",
			Price:       38000,
			Description: "Created by empowered barista",
			Photo:       "http://example.com/barista-blend.jpg",
			IsAvailable: &isAvail,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", bodyBytes, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("Customer forbidden from category creation", func(t *testing.T) {
		token := GenerateTestToken(2, "customer@coffe.com", "user")

		reqBody := category.CreateCategoryRequest{
			Name: "Customer Attempt Category",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", bodyBytes, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Returns 500 Internal Server Error when permission lookup fails", func(t *testing.T) {
		// Non-existent role ID not in Redis and account-service unreachable
		token := generateCustomToken(99, "unknown_role@coffe.com", 9999, nil)
		if lib.RedisClient != nil {
			lib.RedisClient.Del(context.Background(), "auth:role_permissions:9999")
		}

		reqBody := category.CreateCategoryRequest{
			Name: "Should Fail With 500",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", bodyBytes, token)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
