package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance with custom password and PIN validators.
func New() *Validator {
	v := &Validator{
		validate: validator.New(),
	}

	// Register custom validator here if neccesary
	// v.validate.RegisterValidation("tag_name", validation_fn)

	return v
}

// Struct validates a struct and returns a ValidationErrors map.
func (v *Validator) Struct(s any) error {
	return v.validate.Struct(s)
}

// ValidationErrors returns a readable error message for validation errors.
func ValidationErrors(err error) string {
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err.Error()
	}

	validationErrors := err.(validator.ValidationErrors)
	if len(validationErrors) == 0 {
		return ""
	}

	// Return first error for simplicity
	for _, e := range validationErrors {
		return FormatFieldError(e)
	}
	return "validation failed"
}

// FormatFieldError formats a single validation error.
func FormatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + e.Param() + " characters"
	case "max":
		return field + " must be at most " + e.Param() + " characters"
	case "len":
		return field + " must be exactly " + e.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "numeric":
		return field + " must contain only digits"
	case "uri":
		return field + " must be a valid URL"
	default:
		return field + " is invalid (" + tag + ")"
	}
}
