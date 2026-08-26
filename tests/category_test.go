package tests

import (
	"encoding/json"
	"io"
	"testing"
)

type categoryItem struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type getPaginatedCategoriesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Data        []categoryItem `json:"data"`
		TotalData   int            `json:"totalData"`
		TotalPages  int            `json:"totalPages"`
		CurrentPage int            `json:"currentPage"`
		PageSize    int            `json:"pageSize"`
		LastPage    bool           `json:"lastPage"`
	} `json:"data"`
}

type createCategoryResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    categoryItem `json:"data"`
}

type genericCategoryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

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

		respBody, _ := io.ReadAll(resp.Body)
		var res getPaginatedCategoriesResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message != "Success" {
			t.Errorf("Expected message 'Success', got '%s'", res.Message)
		}
		if len(res.Data.Data) == 0 {
			t.Fatalf("Expected categories data array in pagination to be non-empty")
		}

		// Assert index 0 category item details
		firstCat := res.Data.Data[0]
		if firstCat.Id <= 0 {
			t.Errorf("Expected valid category ID for index 0, got %d", firstCat.Id)
		}
		if firstCat.Name == "" {
			t.Errorf("Expected non-empty category name for index 0")
		}
	})

	t.Run("POST /categories - Admin Create Category 201 + Direct DB Assertion", func(t *testing.T) {
		body := []byte(`{"name": "Non-Coffee & Tea"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 201 Created, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res createCategoryResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message != "Category created successfully" {
			t.Errorf("Expected message 'Category created successfully', got '%s'", res.Message)
		}
		if res.Data.Id <= 0 {
			t.Errorf("Expected valid category ID in data.id, got %d", res.Data.Id)
		}
		if res.Data.Name != "Non-Coffee & Tea" {
			t.Errorf("Expected category name 'Non-Coffee & Tea' in data.name, got '%s'", res.Data.Name)
		}

		// Direct DB State Assertion using returned ID
		var name string
		err = dbConn.QueryRow(`SELECT name FROM tm_categories WHERE id = $1`, res.Data.Id).Scan(&name)
		if err != nil || name != "Non-Coffee & Tea" {
			t.Fatalf("Direct DB Verification Failed: Category ID %d with name 'Non-Coffee & Tea' not found in PostgreSQL DB!", res.Data.Id)
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

	t.Run("POST /categories - Unauthorized Without Token 401", func(t *testing.T) {
		body := []byte(`{"name": "Unauthorized Category"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /categories - Forbidden for Customer Role 403", func(t *testing.T) {
		body := []byte(`{"name": "Forbidden Category"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/categories", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			t.Fatalf("Expected HTTP 403 Forbidden, got %v", resp.StatusCode)
		}
	})

	t.Run("DELETE /categories - Delete Category 200 + Direct DB Assertion", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories?id=10", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericCategoryResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB State Assertion
		var count int
		err = dbConn.QueryRow(`SELECT COUNT(*) FROM tm_categories WHERE id = 10`).Scan(&count)
		if err != nil || count != 0 {
			t.Fatalf("Direct DB Verification Failed: Category ID 10 still exists in DB!")
		}
	})

	t.Run("DELETE /categories - Category Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})

	t.Run("DELETE /categories - Missing ID 400", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/categories", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			t.Fatalf("Expected HTTP 400 Bad Request, got %v", resp.StatusCode)
		}
	})
}
