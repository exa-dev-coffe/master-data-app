package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func TestUploadSuite(t *testing.T) {
	dbConn, teardown := SetupTestPostgres(t)
	defer teardown()

	_, minioTeardown := SetupTestMinio(t)
	defer minioTeardown()

	app := SetupTestApp(dbConn)
	adminToken := GenerateTestToken(1, "admin@test.com", "admin")
	customerToken := GenerateTestToken(100, "customer@test.com", "customer")

	t.Run("POST /upload/upload-menu - Real MinIO Upload Success 200", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="file"; filename="test_menu.jpg"`)
		h.Set("Content-Type", "image/jpeg")
		part, err := writer.CreatePart(h)
		if err != nil {
			t.Fatalf("Failed to create form part: %v", err)
		}

		// Valid JPEG magic header bytes
		jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48}
		_, _ = part.Write(jpegHeader)
		_ = writer.Close()

		req := httptest.NewRequest("POST", "/api/1.0/upload/upload-menu", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+adminToken)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK from MinIO Testcontainer, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /upload/upload-menu - Missing File 400", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/upload/upload-menu", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /upload/delete-menu - Real MinIO Delete Photo 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/upload/delete-menu?url=https://storage.eka-dev.cloud/project/coffe/images/menus/test_menu.jpg", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK from MinIO Testcontainer, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("DELETE /upload/delete-menu - Invalid URL Format 400", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "DELETE", "/api/1.0/upload/delete-menu?url=http://minio/invalid/path/photo.jpg", nil, adminToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request for invalid URL format, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("POST /upload/upload-menu - Unauthorized Without Token 401", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/upload/upload-menu", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /upload/upload-menu - Forbidden for Customer Role 403", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "POST", "/api/1.0/upload/upload-menu", nil, customerToken)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 403 {
			t.Fatalf("Expected HTTP 403 Forbidden, got %v", resp.StatusCode)
		}
	})
}
