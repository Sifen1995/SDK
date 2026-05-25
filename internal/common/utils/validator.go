package utils

import "github.com/go-playground/validator/v10"

var validate = validator.New()

// ValidateStruct validates structural tags of any given struct.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
