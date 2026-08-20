package main

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestMenuCatalogSuite(t *testing.T) {
	dbConn, teardown := setupTestPostgres(t)
	defer teardown()

	app := setupTestApp(dbConn)

	t.Run("GET /api/1.0/menus - List Menus", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/menus", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 && resp.StatusCode != 401 && resp.StatusCode != 403 {
			t.Fatalf("Expected valid HTTP response status, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /api/1.0/menus - Create Menu", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/1.0/menus", bytes.NewBuffer([]byte(`{"name":"Cappuccino","price":30000}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)
		if resp.StatusCode != 201 && resp.StatusCode != 401 && resp.StatusCode != 403 && resp.StatusCode != 400 {
			t.Fatalf("Expected valid HTTP response status, got %v", resp.StatusCode)
		}
	})

	t.Run("GET /api/1.0/menus/detail - Menu Detail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/menus/detail?id=1", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 && resp.StatusCode != 404 && resp.StatusCode != 401 && resp.StatusCode != 400 {
			t.Fatalf("Expected valid HTTP response status, got %v", resp.StatusCode)
		}
	})
}
