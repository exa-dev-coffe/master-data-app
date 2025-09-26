package lib

import (
	"errors"
	"mime/multipart"

	"eka-dev.com/master-data/utils/response"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func formatValidationError(err error) map[string]string {
	errorsMap := make(map[string]string)
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		for _, e := range errs {
			fieldName := e.Field() // default pakai nama struct field
			// ambil nama dari json tag kalau ada
			if jsonTag := e.StructField(); jsonTag != "" {
				fieldName = e.Field()
			}
			errorsMap[fieldName] = validationMessage(e)
		}
	}
	return errorsMap
}

func validateStruct(s interface{}) error {
	return Validate.Struct(s)
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return "must be at least " + e.Param() + " characters"
	case "max":
		return "must be at most " + e.Param() + " characters"
	default:
		return "failed on " + e.Tag()
	}
}

func ValidateRequest(s interface{}) error {
	err := validateStruct(s)
	if err != nil {
		return response.BadRequest("Validation error", formatValidationError(err))
	}
	return nil
}

func ValidateImageFile(fileHeader *multipart.FileHeader) (*multipart.FileHeader, error) {
	if fileHeader == nil {
		return nil, response.BadRequest("File is required", nil)
	}

	// Check file size (e.g., max 5MB)
	const maxFileSize = 5 * 1024 * 1024 // 5MB
	if fileHeader.Size > maxFileSize {
		return nil, response.BadRequest("File size exceeds the maximum limit of 5MB", nil)
	}

	// Check file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowedTypes[fileHeader.Header.Get("Content-Type")] {
		return nil, response.BadRequest("Invalid file type. Only JPEG, PNG, and GIF are allowed", nil)
	}

	return fileHeader, nil
}
