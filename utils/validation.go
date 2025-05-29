package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError mengubah validation errors jadi map field => pesan error
func FormatValidationError(errs validator.ValidationErrors) map[string]string {
	errMap := make(map[string]string)

	for _, err := range errs {
		field := err.Field()
		tag := err.Tag()

		var msg string
		switch tag {
		case "required":
			msg = fmt.Sprintf("%s wajib diisi", field)
		case "email":
			msg = fmt.Sprintf("%s harus berupa email yang valid", field)
		case "uuid4":
			msg = fmt.Sprintf("%s harus berupa UUID v4", field)
		case "numeric":
			msg = fmt.Sprintf("%s harus berupa angka", field)
		case "birthdate":
			msg = fmt.Sprintf("%s tidak boleh di masa depan", field)
		default:
			msg = fmt.Sprintf("%s tidak valid", field)
		}

		errMap[field] = msg
	}

	return errMap
}
