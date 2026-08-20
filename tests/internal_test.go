package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func executeInternalTestRequest(app *fiber.App, method, url, query string, customTimestamp ...string) (*http.Response, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	if len(customTimestamp) > 0 {
		timestamp = customTimestamp[0]
	}
	signature := GenerateHMACSignature(query, "", timestamp)

	fullURL := url
	if query != "" {
		fullURL = fmt.Sprintf("%s?%s", url, query)
	}

	req := httptest.NewRequest(method, fullURL, bytes.NewBuffer(nil))
	req.Header.Set("X-Signature", signature)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")

	return app.Test(req)
}

func TestInternalSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)

	// Seed test data in PostgreSQL DB
	_, err := dbConn.Exec(`
		INSERT INTO tm_tables (id, name) VALUES (1, 'Table 1') ON CONFLICT (id) DO NOTHING;
		INSERT INTO tm_menus (id, name, description, price, is_available, photo)
		VALUES (10, 'Espresso', 'Dark roast coffee', 25000.00, true, 'coffee.jpg') ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed internal test data: %v", err)
	}

	t.Run("GET /internal/available-menus-table - Internal Available Menus & Table 200", func(t *testing.T) {
		query := "ids=10&tableId=1"
		resp, err := executeInternalTestRequest(app, "GET", "/api/internal/available-menus-table", query)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /internal/data-menus-table - Internal Data Menus & Tables List 200", func(t *testing.T) {
		query := "ids=10&tableIds=1"
		resp, err := executeInternalTestRequest(app, "GET", "/api/internal/data-menus-table", query)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /internal/available-menus-table - Missing HMAC Headers 401", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/internal/available-menus-table?ids=10&tableId=1", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("GET /internal/available-menus-table - Expired Timestamp Replay Attack 401", func(t *testing.T) {
		expiredTime := time.Now().Add(-10 * time.Minute).UTC().Format(time.RFC3339)
		query := "ids=10&tableId=1"
		resp, err := executeInternalTestRequest(app, "GET", "/api/internal/available-menus-table", query, expiredTime)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 401 Unauthorized for expired timestamp, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /internal/available-menus-table - Nonexistent Table 400", func(t *testing.T) {
		query := "ids=10&tableId=99999"
		resp, err := executeInternalTestRequest(app, "GET", "/api/internal/available-menus-table", query)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request for Nonexistent Table, got %v: %s", resp.StatusCode, string(respBody))
		}
	})
}
