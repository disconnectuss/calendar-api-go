package common

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Validator = validator.Validate

func NewValidator() *Validator {
	v := validator.New()
	_ = v.RegisterValidation("rfc3339", validateRFC3339)
	return v
}

func validateRFC3339(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String())
	return err == nil
}
