package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// InitValidator inialisasi validator
func InitValidator() {
	validate = validator.New()

	// Register custom validations
	registerCustomValidations()
}

// ValidateStruct validasi struct dengan custom rules
func ValidateStruct(s interface{}) error {
	if validate == nil {
		InitValidator()
	}

	if err := validate.Struct(s); err != nil {
		// Format error messages lebih friendly
		return formatValidationErrors(err)
	}

	return nil
}

func registerCustomValidations() {
	// 1. Valid Role - untuk user management
	validate.RegisterValidation("valid_role", func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		validRoles := map[string]bool{
			"super_admin": true,
			"admin":       true,
			"staff":       true,
		}
		return validRoles[role]
	})

	// 2. Password strength - untuk user registration/update
	validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 6 {
			return false
		}
		return true
	})

	// 3. Product Code format - untuk products table
	validate.RegisterValidation("product_code", func(fl validator.FieldLevel) bool {
		code := fl.Field().String()
		// Format: PROD-001, ITEM-123, dll (bisa diatur)
		if len(code) < 3 || len(code) > 50 {
			return false
		}
		return true
	})

	// 4. Positive Number - untuk price, stock, quantity
	validate.RegisterValidation("positive", func(fl validator.FieldLevel) bool {
		switch fl.Field().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fl.Field().Int() > 0
		case reflect.Float32, reflect.Float64:
			return fl.Field().Float() > 0
		case reflect.String:
			// Untuk string yang harus konvert ke number dulu
			return true
		default:
			return false
		}
	})

	// 5. UUID format - untuk foreign key references (RELEVAN)
	validate.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		uuidStr := fl.Field().String()
		pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
		matched, _ := regexp.MatchString(pattern, strings.ToLower(uuidStr))
		return matched
	})
}

// formatValidationErrors konversi error validator ke format yang lebih readable
func formatValidationErrors(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)

		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()

			// Custom error messages berdasarkan tag
			switch tag {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errors[field] = fmt.Sprintf("%s must be a valid email", field)
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be maximum %s characters", field, e.Param())
			case "valid_role":
				errors[field] = fmt.Sprintf("%s must be one of: super_admin, admin, staff", field)
			case "strong_password":
				errors[field] = fmt.Sprintf("%s must be at least 6 characters", field)
			case "uuid4":
				errors[field] = fmt.Sprintf("%s must be a valid UUID v4", field)
			case "positive":
				errors[field] = fmt.Sprintf("%s must be positive number", field)
			default:
				errors[field] = fmt.Sprintf("%s failed %s validation", field, tag)
			}
		}

		return fmt.Errorf("validation failed: %v", errors)
	}

	return err
}

// RegisterCustomValidation tambah custom validation rule baru
func RegisterCustomValidation(tag string, fn validator.Func) error {
	if validate == nil {
		InitValidator()
	}

	return validate.RegisterValidation(tag, fn)
}
