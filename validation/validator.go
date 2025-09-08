package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validator represents a custom validator
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("uuid", validateUUID)
	v.RegisterValidation("uuid4", validateUUID4)
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("id_number", validateIDNumber)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("username", validateUsername)

	// Register custom tag name function
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validator: v,
	}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validator.Var(field, tag)
}

// GetValidationErrors returns detailed validation errors
func (v *Validator) GetValidationErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	if err == nil {
		return validationErrors
	}

	// Check if it's a validation error
	if validationErr, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErr {
			validationErrors = append(validationErrors, ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   e.Value(),
				Message: getErrorMessage(e),
			})
		}
	}

	return validationErrors
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "numeric":
		return fmt.Sprintf("%s must be numeric", fe.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fe.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", fe.Field())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", fe.Field())
	case "id_number":
		return fmt.Sprintf("%s must be a valid ID number", fe.Field())
	case "password":
		return fmt.Sprintf("%s must be a strong password", fe.Field())
	case "username":
		return fmt.Sprintf("%s must be a valid username", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is not valid", fe.Field())
	}
}

// Custom validators

// validateUUID validates if a string is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	uuidStr := fl.Field().String()
	if uuidStr == "" {
		return true // Let required validator handle empty values
	}
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

// validateUUID4 validates if a string is a valid UUID v4
func validateUUID4(fl validator.FieldLevel) bool {
	uuidStr := fl.Field().String()
	if uuidStr == "" {
		return true // Let required validator handle empty values
	}
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return false
	}
	return parsedUUID.Version() == 4
}

// validatePhone validates Indonesian phone numbers
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true // Let required validator handle empty values
	}

	// Remove spaces and special characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	// Check if it starts with +62 or 0
	if strings.HasPrefix(phone, "+62") {
		phone = "0" + phone[3:]
	}

	// Check if it starts with 0 and has 10-13 digits
	if !strings.HasPrefix(phone, "0") {
		return false
	}

	// Check if it contains only digits
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Check length (10-13 digits including the leading 0)
	return len(phone) >= 10 && len(phone) <= 13
}

// validateIDNumber validates Indonesian ID numbers (KTP)
func validateIDNumber(fl validator.FieldLevel) bool {
	idNumber := fl.Field().String()
	if idNumber == "" {
		return true // Let required validator handle empty values
	}

	// Remove spaces and special characters
	idNumber = strings.ReplaceAll(idNumber, " ", "")
	idNumber = strings.ReplaceAll(idNumber, ".", "")

	// Check if it contains only digits
	for _, char := range idNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Indonesian KTP is 16 digits
	return len(idNumber) == 16
}

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true // Let required validator handle empty values
	}

	// At least 8 characters
	if len(password) < 8 {
		return false
	}

	// At least one uppercase letter
	hasUpper := false
	// At least one lowercase letter
	hasLower := false
	// At least one digit
	hasDigit := false
	// At least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 126: // Printable ASCII special characters
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if username == "" {
		return true // Let required validator handle empty values
	}

	// Length between 3 and 20 characters
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// Must start with a letter
	if username[0] < 'a' || username[0] > 'z' {
		if username[0] < 'A' || username[0] > 'Z' {
			return false
		}
	}

	// Can contain letters, numbers, underscores, and hyphens
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(err error) *ValidationErrorResponse {
	validator := NewValidator()
	validationErrors := validator.GetValidationErrors(err)

	return &ValidationErrorResponse{
		Error:   "Validation failed",
		Details: validationErrors,
	}
}
