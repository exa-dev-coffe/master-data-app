package tests

import (
	"encoding/json"
	"io"
	"testing"
)

type menuItem struct {
	Id           int64   `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	IsAvailable  bool    `json:"isAvailable"`
	Photo        string  `json:"photo"`
	CategoryId   int64   `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
}

type getPaginatedMenusResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Data        []menuItem `json:"data"`
		TotalData   int        `json:"totalData"`
		TotalPages  int        `json:"totalPages"`
		CurrentPage int        `json:"currentPage"`
		PageSize    int        `json:"pageSize"`
		LastPage    bool       `json:"lastPage"`
	} `json:"data"`
}

type getListMenusResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    []menuItem `json:"data"`
}

type getMenuDetailResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    menuItem `json:"data"`
}

type genericMenuResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestMenuSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")
	baristaToken := GenerateTestToken(2, "barista@test.com", "barista")
	customerToken := GenerateTestToken(100, "customer@test.com", "customer")

	// Seed test data into PostgreSQL DB
	_, err := dbConn.Exec(`
		INSERT INTO tm_categories (id, name) VALUES (1, 'Espresso Base') ON CONFLICT (id) DO NOTHING;
		INSERT INTO tm_menus (id, name, description, price, is_available, photo, category_id)
		VALUES (10, 'Espresso', 'Dark roast coffee', 25000.00, true, 'https://storage.eka-dev.cloud/project/coffe/coffee.jpg', 1) ON CONFLICT (id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed menu test data: %v", err)
	}

	t.Run("GET /menus - List Menus 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res getPaginatedMenusResponse
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
			t.Fatalf("Expected menus data array in pagination to be non-empty")
		}

		// Assert index 0 menu item details
		firstMenu := res.Data.Data[0]
		if firstMenu.Id <= 0 {
			t.Errorf("Expected valid menu ID for index 0, got %d", firstMenu.Id)
		}
		if firstMenu.Name == "" {
			t.Errorf("Expected non-empty menu name for index 0")
		}
		if firstMenu.Price <= 0 {
			t.Errorf("Expected menu price > 0, got %f", firstMenu.Price)
		}
	})

	t.Run("GET /menus/detail - Menu Detail 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/detail?id=10", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res getMenuDetailResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message != "Success" {
			t.Errorf("Expected message 'Success', got '%s'", res.Message)
		}
		if res.Data.Id != 10 {
			t.Errorf("Expected menu ID 10, got %d", res.Data.Id)
		}
		if res.Data.Name != "Espresso" {
			t.Errorf("Expected menu name 'Espresso', got '%s'", res.Data.Name)
		}
		if res.Data.Price != 25000.00 {
			t.Errorf("Expected menu price 25000.00, got %f", res.Data.Price)
		}
	})

	t.Run("GET /menus/detail - Menu Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/menus/detail?id=99999", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /menus - Admin Create Menu 201 + Direct DB Assertion (All Columns)", func(t *testing.T) {
		body := []byte(`{
			"name": "Matcha Latte",
			"description": "Premium Japanese Matcha",
			"price": 35000.00,
			"isAvailable": true,
			"photo": "https://storage.eka-dev.cloud/project/coffe/matcha.jpg",
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

		respBody, _ := io.ReadAll(resp.Body)
		var res genericMenuResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification across ALL inserted columns
		var (
			name        string
			description string
			price       float64
			isAvailable bool
			photo       string
			categoryID  int64
		)
		err = dbConn.QueryRow(`SELECT name, description, price, is_available, photo, category_id FROM tm_menus WHERE name = 'Matcha Latte'`).Scan(&name, &description, &price, &isAvailable, &photo, &categoryID)
		if err != nil {
			t.Fatalf("Direct DB Verification Failed: Menu 'Matcha Latte' not found in DB: %v", err)
		}
		if name != "Matcha Latte" {
			t.Errorf("Expected name 'Matcha Latte', got '%s'", name)
		}
		if description != "Premium Japanese Matcha" {
			t.Errorf("Expected description 'Premium Japanese Matcha', got '%s'", description)
		}
		if price != 35000.00 {
			t.Errorf("Expected price 35000.00, got %f", price)
		}
		if !isAvailable {
			t.Errorf("Expected is_available true, got false")
		}
		if photo != "https://storage.eka-dev.cloud/project/coffe/matcha.jpg" {
			t.Errorf("Expected photo URL match, got '%s'", photo)
		}
		if categoryID != 1 {
			t.Errorf("Expected category_id 1, got %d", categoryID)
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
		body := []byte(`{
			"name": "Forbidden Menu",
			"description": "Desc",
			"price": 20000.00,
			"isAvailable": true,
			"photo": "https://storage.eka-dev.cloud/project/coffe/photo.jpg"
		}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/menus", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			t.Fatalf("Expected HTTP 403 Forbidden, got %v", resp.StatusCode)
		}
	})

	t.Run("PUT /menus - Admin Update Menu 200 + Direct DB Assertion (All Columns)", func(t *testing.T) {
		body := []byte(`{
			"id": 10,
			"name": "Updated Espresso Special",
			"description": "Extra Shot Coffee Description",
			"price": 32000.00,
			"isAvailable": false,
			"photo": "https://storage.eka-dev.cloud/project/coffe/photo_updated.jpg",
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

		respBody, _ := io.ReadAll(resp.Body)
		var res genericMenuResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification across ALL updated columns
		var (
			name        string
			description string
			price       float64
			isAvailable bool
			photo       string
			categoryID  int64
		)
		err = dbConn.QueryRow(`SELECT name, description, price, is_available, photo, category_id FROM tm_menus WHERE id = 10`).Scan(&name, &description, &price, &isAvailable, &photo, &categoryID)
		if err != nil {
			t.Fatalf("Direct DB Verification Failed for Menu ID 10: %v", err)
		}
		if name != "Updated Espresso Special" {
			t.Errorf("Expected updated name 'Updated Espresso Special', got '%s'", name)
		}
		if description != "Extra Shot Coffee Description" {
			t.Errorf("Expected updated description 'Extra Shot Coffee Description', got '%s'", description)
		}
		if price != 32000.00 {
			t.Errorf("Expected updated price 32000.00, got %f", price)
		}
		if isAvailable != false {
			t.Errorf("Expected updated is_available false, got true")
		}
		if photo != "https://storage.eka-dev.cloud/project/coffe/photo_updated.jpg" {
			t.Errorf("Expected updated photo URL, got '%s'", photo)
		}
		if categoryID != 1 {
			t.Errorf("Expected updated category_id 1, got %d", categoryID)
		}
	})

	t.Run("PUT /menus - Menu Not Found 404", func(t *testing.T) {
		body := []byte(`{
			"id": 99999,
			"name": "Nonexistent Menu",
			"description": "Desc",
			"price": 10000.00,
			"isAvailable": true,
			"photo": "https://storage.eka-dev.cloud/project/coffe/photo.jpg"
		}`)
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/menus", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
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

		respBody, _ := io.ReadAll(resp.Body)
		var res getPaginatedMenusResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message != "Success" {
			t.Errorf("Expected message 'Success', got '%s'", res.Message)
		}
	})

	t.Run("PATCH /menus/set-category - Admin Set Menu Category 200 + Direct DB Assertion", func(t *testing.T) {
		body := []byte(`{
			"id": 10,
			"categoryId": 1
		}`)
		resp, err := ExecuteTestRequest(app, "PATCH", "/api/1.0/menus/set-category", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericMenuResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification
		var categoryID int64
		err = dbConn.QueryRow(`SELECT category_id FROM tm_menus WHERE id = 10`).Scan(&categoryID)
		if err != nil || categoryID != 1 {
			t.Fatalf("Direct DB Verification Failed: Expected category_id 1 for Menu ID 10, got %d (err: %v)", categoryID, err)
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

		respBody, _ := io.ReadAll(resp.Body)
		var res getListMenusResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message != "Success" {
			t.Errorf("Expected message 'Success', got '%s'", res.Message)
		}
		if len(res.Data) == 0 {
			t.Fatalf("Expected menus by category data array to be non-empty")
		}
		if res.Data[0].Id <= 0 {
			t.Errorf("Expected valid menu ID > 0 for index 0 in category 1, got %d", res.Data[0].Id)
		}
	})

	t.Run("PATCH /menus/availability - Barista Toggle Menu Availability 200 + Direct DB Assertion", func(t *testing.T) {
		body := []byte(`{
			"id": 10,
			"isAvailable": false
		}`)
		resp, err := ExecuteTestRequest(app, "PATCH", "/api/1.0/menus/availability", body, baristaToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericMenuResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification
		var isAvailable bool
		err = dbConn.QueryRow(`SELECT is_available FROM tm_menus WHERE id = 10`).Scan(&isAvailable)
		if err != nil || isAvailable != false {
			t.Fatalf("Direct DB Verification Failed: Menu ID 10 is_available flag was not updated to false in DB!")
		}
	})

	t.Run("DELETE /menus - Admin Soft Delete Menu 200 + Direct DB Assertion", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/menus?id=10", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericMenuResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification
		var isDeleted bool
		err = dbConn.QueryRow(`SELECT is_deleted FROM tm_menus WHERE id = 10`).Scan(&isDeleted)
		if err != nil || !isDeleted {
			t.Fatalf("Direct DB Verification Failed: Menu ID 10 is_deleted flag is not true in DB!")
		}
	})

	t.Run("DELETE /menus - Menu Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/menus?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})
}
