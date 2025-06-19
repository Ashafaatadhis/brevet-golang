package validators

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
		case "group_type":
			msg = fmt.Sprintf("%s harus salah satu dari: mahasiswa_gunadarma, mahasiswa_non_gunadarma, umum", field)
		case "role_type":
			msg = fmt.Sprintf("%s harus salah satu dari: siswa, guru, admin", field)
		case "day_type":
			msg = fmt.Sprintf("%s harus salah satu dari: Monday, Tuesday, ..., Sunday", field)
		case "course_type":
			msg = fmt.Sprintf("%s harus salah satu dari: online, offline", field)
		default:
			msg = fmt.Sprintf("%s tidak valid", field)
		}

		errMap[field] = msg
	}

	return errMap
}
