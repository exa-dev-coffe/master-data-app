package tests

import (
	"io"
	"testing"
)

func TestCategorySuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")
	customerToken := GenerateTestToken(100, "customer@test.com", "customer")

	// Seed test category into PostgreSQL DB
	_, err := dbConn.Exec(`
		INSERT INTO tm_categories (id, name) VALUES (10, 'Coffee & Espresso') ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed category test data: %v", err)
	}

	t.Run("GET /categories - List Categories 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/categories", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /categories - Admin Create Category 201", func(t *testing.T) {
		body := []byte(`{"name": "Non-Coffee & Tea"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 201 Created, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /categories - Empty Body Validation Error 400", func(t *testing.T) {
		body := []byte(`{}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /categories - Delete Category 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories?id=10", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /categories - Category Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /categories - Missing ID 400", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /categories - Unauthorized without Token 401", func(t *testing.T) {
		body := []byte(`{"name": "Snacks"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /categories - Forbidden for Customer Role 403", func(t *testing.T) {
		body := []byte(`{"name": "Snacks"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 403 Forbidden, got %v: %s", resp.StatusCode, string(respBody))
		}
	})
}
