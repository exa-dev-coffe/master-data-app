package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"eka-dev.cloud/master-data/modules/menu"
	"eka-dev.cloud/master-data/modules/promotion"
	"eka-dev.cloud/master-data/utils/response"
)

type createPromoResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    struct {
		ID int64 `json:"id"`
	} `json:"data"`
}

type getMenusPaginatedResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    response.Pagination[[]menu.Menu] `json:"data"`
}

type getPromotionsPaginatedResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    response.Pagination[[]promotion.Promotion] `json:"data"`
}

func TestPromotionModule(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	token := GenerateTestToken(1, "admin@coffe.com", "ADMIN")

	// Seed category and menu for testing
	var categoryID int64
	err := dbConn.QueryRow("INSERT INTO tm_categories (name) VALUES ('Espresso Based') RETURNING id").Scan(&categoryID)
	if err != nil {
		t.Fatalf("Failed to seed category: %v", err)
	}

	var menuID int64
	err = dbConn.QueryRow(`
		INSERT INTO tm_menus (name, description, price, is_available, category_id, photo) 
		VALUES ('Americano Test', 'Freshly brewed espresso', 25000.00, true, $1, 'http://example.com/americano.jpg') 
		RETURNING id
	`, categoryID).Scan(&menuID)
	if err != nil {
		t.Fatalf("Failed to seed menu: %v", err)
	}

	now := time.Now()
	startTime := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := now.Add(24 * time.Hour).Format(time.RFC3339)

	var createdPromoID int64

	t.Run("POST /api/1.0/promotions - Create Product Level Promo (20% OFF)", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Americano 20%% OFF",
			"targetType": "PRODUCT",
			"targetId": %d,
			"discountType": "PERCENTAGE",
			"discountValue": 20.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, menuID, startTime, endTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("Expected status 201, got %d: %s", res.StatusCode, string(body))
		}

		body, _ := io.ReadAll(res.Body)
		var apiRes createPromoResponse
		if err := json.Unmarshal(body, &apiRes); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		createdPromoID = apiRes.Data.ID
		if createdPromoID == 0 {
			t.Errorf("Expected valid promotion ID, got 0")
		}
	})

	t.Run("GET /api/1.0/menus - Verify Menu Effective Price with Product Promo", func(t *testing.T) {
		res, err := ExecuteTestRequest(app, http.MethodGet, "/api/1.0/menus?page=1&size=10", nil, token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		var apiRes getMenusPaginatedResponse
		if err := json.Unmarshal(body, &apiRes); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		var testMenu *menu.Menu
		for _, m := range apiRes.Data.Data {
			if m.Id == menuID {
				testMenu = &m
				break
			}
		}

		if testMenu == nil {
			t.Fatalf("Menu ID %d not found in list response", menuID)
		}

		// Price is 25000. 20% discount = 5000 savings. Effective price = 20000
		expectedEffective := 20000.00
		if testMenu.EffectivePrice != expectedEffective {
			t.Errorf("Expected effectivePrice %f, got %f", expectedEffective, testMenu.EffectivePrice)
		}

		if testMenu.Discount == nil {
			t.Fatalf("Expected Discount detail to be populated")
		}
		if testMenu.Discount.Savings != 5000.00 {
			t.Errorf("Expected savings 5000.00, got %f", testMenu.Discount.Savings)
		}
	})

	t.Run("GET /api/1.0/menus/detail - Verify Single Menu Detail Effective Price and Discount", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/menus/detail?id=%d", menuID)
		res, err := ExecuteTestRequest(app, http.MethodGet, url, nil, token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		var apiRes struct {
			Message string    `json:"message"`
			Success bool      `json:"success"`
			Data    menu.Menu `json:"data"`
		}
		if err := json.Unmarshal(body, &apiRes); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if apiRes.Data.EffectivePrice != 20000.00 {
			t.Errorf("Expected single menu detail effectivePrice 20000.00, got %f", apiRes.Data.EffectivePrice)
		}
		if apiRes.Data.Discount == nil {
			t.Fatalf("Expected single menu detail Discount object to be non-nil")
		}
		if apiRes.Data.Discount.Savings != 5000.00 {
			t.Errorf("Expected savings 5000.00, got %f", apiRes.Data.Discount.Savings)
		}
	})

	t.Run("GET /api/1.0/promotions - List Promotions", func(t *testing.T) {
		res, err := ExecuteTestRequest(app, http.MethodGet, "/api/1.0/promotions?page=1&size=10", nil, token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		body, _ := io.ReadAll(res.Body)
		var apiRes getPromotionsPaginatedResponse
		if err := json.Unmarshal(body, &apiRes); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(apiRes.Data.Data) == 0 {
			t.Errorf("Expected promotions list to contain at least 1 item")
		}
	})

	t.Run("PATCH /api/1.0/promotions/:id/status - Deactivate Promotion", func(t *testing.T) {
		reqBody := `{"isActive": false}`
		url := fmt.Sprintf("/api/1.0/promotions/%d/status", createdPromoID)
		res, err := ExecuteTestRequest(app, http.MethodPatch, url, []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}
	})

	t.Run("GET /api/1.0/menus - Verify Menu Price Returns to Base Price when Promo Inactive", func(t *testing.T) {
		res, err := ExecuteTestRequest(app, http.MethodGet, "/api/1.0/menus?page=1&size=10", nil, token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		body, _ := io.ReadAll(res.Body)
		var apiRes getMenusPaginatedResponse
		if err := json.Unmarshal(body, &apiRes); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		var testMenu *menu.Menu
		for _, m := range apiRes.Data.Data {
			if m.Id == menuID {
				testMenu = &m
				break
			}
		}

		if testMenu.EffectivePrice != 25000.00 {
			t.Errorf("Expected effectivePrice to revert to base price 25000.00, got %f", testMenu.EffectivePrice)
		}
	})

	t.Run("PATCH /api/1.0/promotions/:id/status - Reject Activating Expired Promotion", func(t *testing.T) {
		var expiredPromoID int64
		err := dbConn.QueryRow(`
			INSERT INTO tm_promotions (name, target_type, discount_type, discount_value, start_at, end_at, is_active)
			VALUES ('Expired Promo Test', 'ALL', 'PERCENTAGE', 10.0, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day', false)
			RETURNING id
		`).Scan(&expiredPromoID)
		if err != nil {
			t.Fatalf("Failed to seed expired promo: %v", err)
		}

		reqBody := `{"isActive": true}`
		url := fmt.Sprintf("/api/1.0/promotions/%d/status", expiredPromoID)
		res, err := ExecuteTestRequest(app, http.MethodPatch, url, []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /api/1.0/promotions - Reject Percentage > 100", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Invalid 150%% Promo",
			"targetType": "ALL",
			"discountType": "PERCENTAGE",
			"discountValue": 150.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, startTime, endTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /api/1.0/promotions - Reject Missing TargetID for PRODUCT TargetType", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Missing Target Product",
			"targetType": "PRODUCT",
			"discountType": "PERCENTAGE",
			"discountValue": 10.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, startTime, endTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /api/1.0/promotions - Reject End Date Before Start Date", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Invalid Date Range Promo",
			"targetType": "ALL",
			"discountType": "PERCENTAGE",
			"discountValue": 10.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, endTime, startTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400, got %d", res.StatusCode)
		}
	})

	t.Run("POST /api/1.0/promotions - Create Storewide Promo (TargetType ALL)", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Storewide Sale 10%%",
			"targetType": "ALL",
			"discountType": "PERCENTAGE",
			"discountValue": 10.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, startTime, endTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(res.Body)
			t.Fatalf("Expected status 201, got %d: %s", res.StatusCode, string(body))
		}
	})

	t.Run("POST /api/1.0/promotions - Reject Overlapping Same Target Promotion", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{
			"name": "Overlapping Americano Promo",
			"targetType": "PRODUCT",
			"targetId": %d,
			"discountType": "PERCENTAGE",
			"discountValue": 15.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, menuID, startTime, endTime)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400 for overlapping promo, got %d", res.StatusCode)
		}
	})

	t.Run("POST /api/1.0/promotions - Reject End Date Overlapping Existing Promotion", func(t *testing.T) {
		earlyStart := now.Add(-3 * time.Hour).Format(time.RFC3339)
		midEnd := now.Add(1 * time.Hour).Format(time.RFC3339)

		reqBody := fmt.Sprintf(`{
			"name": "End Date Overlap Promo",
			"targetType": "PRODUCT",
			"targetId": %d,
			"discountType": "PERCENTAGE",
			"discountValue": 10.0,
			"startAt": "%s",
			"endAt": "%s"
		}`, menuID, earlyStart, midEnd)

		res, err := ExecuteTestRequest(app, http.MethodPost, "/api/1.0/promotions", []byte(reqBody), token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400 for end-date overlap promo, got %d", res.StatusCode)
		}
	})

	t.Run("DELETE /api/1.0/promotions/:id - Soft Delete Promotion", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/promotions/%d", createdPromoID)
		res, err := ExecuteTestRequest(app, http.MethodDelete, url, nil, token)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}
	})
}
