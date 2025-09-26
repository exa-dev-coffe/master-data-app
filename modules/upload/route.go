package upload

import (
	"eka-dev.com/master-data/lib"
	"eka-dev.com/master-data/middleware"
	"eka-dev.com/master-data/utils/common"
	"eka-dev.com/master-data/utils/response"
	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	UploadMenuFoto(c *fiber.Ctx) error
	DeleteMenuFoto(c *fiber.Ctx) error
}

type handler struct {
	service Service
}

// NewHandler return handler dan daftarin route
func NewHandler(app *fiber.App) Handler {
	service := NewUploadService()
	handler := &handler{service: service}

	// mapping routes
	routes := app.Group("/api/1.0/upload")
	routes.Post("/upload-menu", middleware.RequireRole("admin"), handler.UploadMenuFoto)
	routes.Delete("/delete", middleware.RequireRole("admin"), handler.DeleteMenuFoto)

	return handler
}

func (s *handler) UploadMenuFoto(c *fiber.Ctx) error {
	// Parse the multipart form:
	form, err := c.MultipartForm()
	if err != nil {
		return response.BadRequest("Failed to parse multipart form: "+err.Error(), nil)
	}

	files := form.File["file"]
	if len(files) == 0 {
		return response.BadRequest("No file is uploaded", nil)
	}

	// For simplicity, we handle only the first file

	fileHeader, err := lib.ValidateImageFile(files[0])
	if err != nil {
		return err
	}

	res, err := s.service.UploadMenuFoto(fileHeader)

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("File uploaded successfully", res))
}

func (s *handler) DeleteMenuFoto(c *fiber.Ctx) error {
	var request common.DeleteImageRequest
	err := c.QueryParser(&request)
	if err != nil {
		return response.BadRequest("Invalid query parameters: "+err.Error(), nil)
	}

	err = lib.ValidateRequest(request)
	if err != nil {
		return err
	}

	err = s.service.DeleteMenuFoto(request.Url)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("File deleted successfully", nil))
}
