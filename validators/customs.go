package validators

import (
	"brevet-api/models"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
)

// GroupTypeValidator checks if group_type value is valid
func GroupTypeValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Jika nil, anggap valid (karena tidak wajib)
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return true
		}
	}

	var val string

	// Handle pointer dan non-pointer string
	if field.Kind() == reflect.Ptr {
		val = field.Elem().String()
	} else if field.Kind() == reflect.String {
		val = field.String()
	} else {
		return false
	}

	// Validasi nilai
	switch models.GroupType(val) {
	case models.MahasiswaGunadarma, models.MahasiswaNonGunadarma, models.Umum:
		return true
	default:
		return false
	}
}

// RoleTypeValidator checks if role_type value is valid
func RoleTypeValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Kalau nil, anggap valid (tidak wajib)
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return true
		}
	}

	var val string

	// Ambil nilai string dari pointer atau value biasa
	if field.Kind() == reflect.Ptr {
		val = field.Elem().String()
	} else if field.Kind() == reflect.String {
		val = field.String()
	} else {
		return false
	}

	switch models.RoleType(val) {
	case models.RoleTypeSiswa, models.RoleTypeGuru, models.RoleTypeAdmin:
		return true
	default:
		return false
	}
}

// BirthDateValidator validates that a birth date is not in the future
func BirthDateValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Kalau nil, anggap valid (tidak wajib diisi)
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	var birthDate time.Time
	switch field.Kind() {
	case reflect.Ptr:
		birthDate = field.Elem().Interface().(time.Time)
	case reflect.Struct:
		birthDate = field.Interface().(time.Time)
	default:
		return false
	}

	// Valid jika tanggal lahir tidak setelah hari ini
	return !birthDate.After(time.Now())
}
