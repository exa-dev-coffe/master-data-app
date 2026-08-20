package main

import (
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	dbConn, teardown := setupTestPostgres(t)
	defer teardown()

	app := setupTestApp(dbConn)
	t.Run("GET /health - 200 OK", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200, got %v", resp.StatusCode)
		}
	})
}
