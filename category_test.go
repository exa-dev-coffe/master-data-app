package main

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestCategorySuite(t *testing.T) {
	dbConn, teardown := setupTestPostgres(t)
	defer teardown()

	app := setupTestApp(dbConn)

	t.Run("GET /api/1.0/categories - List Categories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/categories", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 && resp.StatusCode != 401 && resp.StatusCode != 403 {
			t.Fatalf("Expected valid HTTP response status, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /api/1.0/categories - Create Category", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/1.0/categories", bytes.NewBuffer([]byte(`{"name":"Beverage"}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)
		if resp.StatusCode != 201 && resp.StatusCode != 401 && resp.StatusCode != 403 && resp.StatusCode != 400 {
			t.Fatalf("Expected valid HTTP response status, got %v", resp.StatusCode)
		}
	})
}
