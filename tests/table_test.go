package tests

import (
	"fmt"
	"io"
	"testing"
)

func TestTableSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	customerToken := GenerateTestToken(100, "user@test.com", "customer")
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")

	// Seed test table in PostgreSQL DB and get exact returned ID
	var seededTableId int
	err := dbConn.Get(&seededTableId, "INSERT INTO tm_tables (id, name) VALUES (999, 'Table Seed 1') ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name RETURNING id")
	if err != nil {
		t.Fatalf("Failed to seed table test data: %v", err)
	}

	t.Run("GET /tables - Customer List Tables 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/tables?page=1&size=10", nil, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /tables - Admin Create Table 201", func(t *testing.T) {
		body := []byte(`{"name": "VIP Table 1"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 201 Created, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /tables - Empty Body Validation Error 400", func(t *testing.T) {
		body := []byte(`{}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /tables - Forbidden for Customer Role 403", func(t *testing.T) {
		body := []byte(`{"name": "VIP Table 2"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/tables", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 403 Forbidden, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PUT /tables - Admin Update Table 200", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{"id": %d, "name": "Table Seed 1 Updated"}`, seededTableId))
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PUT /tables - Table Not Found 404", func(t *testing.T) {
		body := []byte(`{"id": 99999, "name": "Nonexistent Table"}`)
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /tables - Admin Soft Delete Table 200", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/tables?id=%d", seededTableId)
		resp, err := ExecuteTestRequest(app, "DELETE", url, nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /tables - Table Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/tables?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})
}
