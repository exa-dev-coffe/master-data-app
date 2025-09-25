package utils

import (
	"errors"

	"eka-dev.com/master-data/lib"
	"eka-dev.com/master-data/utils/response"
	"github.com/go-playground/validator/v10"
)

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
	return lib.Validate.Struct(s)
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
