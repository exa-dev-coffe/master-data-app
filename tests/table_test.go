package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type tableItem struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	IsDeleted bool   `json:"isDeleted"`
}

type getPaginatedTablesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Data        []tableItem `json:"data"`
		TotalData   int         `json:"totalData"`
		TotalPages  int         `json:"totalPages"`
		CurrentPage int         `json:"currentPage"`
		PageSize    int         `json:"pageSize"`
		LastPage    bool        `json:"lastPage"`
	} `json:"data"`
}

type genericTableResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestTableSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	app := SetupTestApp(dbConn)
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")
	customerToken := GenerateTestToken(100, "customer@test.com", "customer")

	// Seed test table into PostgreSQL DB dynamically
	var seededTableID int64
	err := dbConn.QueryRow(`
		INSERT INTO tm_tables (name) VALUES ('Meja Seed 10') RETURNING id;
	`).Scan(&seededTableID)
	if err != nil {
		t.Fatalf("Failed to seed table test data: %v", err)
	}

	t.Run("GET /tables - Customer List Tables 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/tables", nil, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res getPaginatedTablesResponse
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
			t.Fatalf("Expected tables data array in pagination to be non-empty")
		}

		// Assert index 0 table item details
		firstTable := res.Data.Data[0]
		if firstTable.Id <= 0 {
			t.Errorf("Expected valid table ID for index 0, got %d", firstTable.Id)
		}
		if firstTable.Name == "" {
			t.Errorf("Expected non-empty table name for index 0")
		}
	})

	t.Run("POST /tables - Admin Create Table 201 + Direct DB Assertion", func(t *testing.T) {
		body := []byte(`{"name": "Meja VIP 99"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 201 Created, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericTableResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}
		if res.Message == "" {
			t.Errorf("Expected non-empty response message")
		}

		// Direct DB Verification
		var name string
		err = dbConn.QueryRow(`SELECT name FROM tm_tables WHERE name = 'Meja VIP 99'`).Scan(&name)
		if err != nil || name != "Meja VIP 99" {
			t.Fatalf("Direct DB Verification Failed: Table 'Meja VIP 99' not found in PostgreSQL DB!")
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
		body := []byte(`{"name": "Forbidden Table"}`)
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/tables", body, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			t.Fatalf("Expected HTTP 403 Forbidden, got %v", resp.StatusCode)
		}
	})

	t.Run("PUT /tables - Admin Update Table 200 + Direct DB Assertion", func(t *testing.T) {
		body := []byte(fmt.Sprintf(`{
			"id": %d,
			"name": "Meja 10 Updated VIP"
		}`, seededTableID))
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericTableResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification
		var name string
		err = dbConn.QueryRow(`SELECT name FROM tm_tables WHERE id = $1`, seededTableID).Scan(&name)
		if err != nil || name != "Meja 10 Updated VIP" {
			t.Fatalf("Direct DB Verification Failed: Table ID %d name was not updated in DB!", seededTableID)
		}
	})

	t.Run("PUT /tables - Table Not Found 404", func(t *testing.T) {
		body := []byte(`{
			"id": 99999,
			"name": "Nonexistent Table"
		}`)
		resp, err := ExecuteTestRequest(app, "PUT", "/api/1.0/tables", body, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})

	t.Run("DELETE /tables - Admin Soft Delete Table 200 + Direct DB Assertion", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/tables?id=%d", seededTableID)
		resp, err := ExecuteTestRequest(app, "DELETE", url, nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK, got %v: %s", resp.StatusCode, string(respBody))
		}

		respBody, _ := io.ReadAll(resp.Body)
		var res genericTableResponse
		if err := json.Unmarshal(respBody, &res); err != nil {
			t.Fatalf("Failed to unmarshal response JSON: %v", err)
		}

		if !res.Success {
			t.Errorf("Expected success to be true, got false")
		}

		// Direct DB Verification
		var isDeleted bool
		err = dbConn.QueryRow(`SELECT is_deleted FROM tm_tables WHERE id = $1`, seededTableID).Scan(&isDeleted)
		if err != nil || !isDeleted {
			t.Fatalf("Direct DB Verification Failed: Table ID %d is_deleted flag is not true in DB!", seededTableID)
		}
	})

	t.Run("DELETE /tables - Table Not Found 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/tables?id=99999", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})
}
