package tests

import (
	"io"
	"testing"
)

func TestMenuSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")
	baristaToken := GenerateTestToken(2, "barista@test.com", "barista")
	customerToken := GenerateTestToken(100, "customer@test.com", "customer")

	// Seed test data in PostgreSQL DB
	_, err := dbConn.Exec(`
		INSERT INTO tm_categories (id, name) VALUES (1, 'Beverages') ON CONFLICT (id) DO NOTHING;
		INSERT INTO tm_menus (id, name, description, price, is_available, photo, category_id)
		VALUES 
			(100, 'Iced Latte', 'Fresh milk with espresso', 28000.00, true, 'http://example.com/coffee.jpg', 1),
			(101, 'Americano', 'Black coffee', 22000.00, true, 'http://example.com/coffee.jpg', NULL)
		ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed menu test data: %v", err)
	}

	t.Run("GET /menus - List Menus 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus?page=1&size=10", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /menus/detail - Menu Detail 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/detail?id=100", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /menus/detail - Menu Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/detail?id=99999", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /menus - Admin Create Menu 201", func(t *testing.T) {
		body := []byte(`{
			"name": "Cappuccino",
			"description": "Rich espresso with steamed milk foam",
			"price": 27000.00,
			"photo": "http://example.com/cappuccino.jpg",
			"isAvailable": true,
			"categoryId": 1
		}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 201 Created, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /menus - Empty Body Validation Error 400", func(t *testing.T) {
		body := []byte(`{}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /menus - Forbidden for Customer Role 403", func(t *testing.T) {
		body := []byte(`{"name": "Test"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 403 Forbidden, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PUT /menus - Admin Update Menu 200", func(t *testing.T) {
		body := []byte(`{
			"id": 100,
			"name": "Iced Latte Deluxe",
			"description": "Premium fresh milk with double espresso",
			"price": 32000.00,
			"photo": "http://example.com/latte.jpg",
			"isAvailable": true,
			"categoryId": 1
		}`)
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/menus", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PUT /menus - Menu Not Found 404", func(t *testing.T) {
		body := []byte(`{
			"id": 99999,
			"name": "Nonexistent Menu",
			"description": "Description text",
			"price": 50000.00,
			"photo": "http://example.com/latte.jpg",
			"isAvailable": true
		}`)
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/menus", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /menus/uncategorized - Admin Get Uncategorized Menus 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/uncategorized", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PATCH /menus/set-category - Admin Set Menu Category 200", func(t *testing.T) {
		body := []byte(`{"id": 101, "categoryId": 1}`)
		resp, err := ExecuteTestRequest(app, "PATCH", "/api/1.0/menus/set-category", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /menus/by-category - Get Menus by Category ID 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/by-category?id=1", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("PATCH /menus/availability - Barista Toggle Menu Availability 200", func(t *testing.T) {
		body := []byte(`{"id": 100, "isAvailable": false}`)
		resp, err := ExecuteTestRequest(app, "PATCH", "/api/1.0/menus/availability", body, baristaToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /menus - Admin Soft Delete Menu 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/menus?id=101", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /menus - Menu Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/menus?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 404 Not Found, got %v: %s", resp.StatusCode, string(respBody))
		}
	})
}
