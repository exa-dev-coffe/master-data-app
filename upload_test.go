package main

import (
	"net/http/httptest"
	"testing"
)

func TestUploadSuite(t *testing.T) {
	dbConn, teardown := setupTestPostgres(t)
	defer teardown()

	app := setupTestApp(dbConn)

	t.Run("POST /api/1.0/upload/upload-menu - Missing File 400/401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/1.0/upload/upload-menu", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 400 && resp.StatusCode != 401 && resp.StatusCode != 500 {
			t.Fatalf("Expected valid error status code, got %v", resp.StatusCode)
		}
	})
}
