package validators

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// BirthDateValidator validates that a birth date is not in the future
func BirthDateValidator(fl validator.FieldLevel) bool {
	birthDate, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return !birthDate.After(time.Now())
}
