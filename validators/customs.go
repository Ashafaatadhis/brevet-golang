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

	var val string

	// Handle pointer case
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return false // required, jadi nil tidak valid
		}
		val = field.Elem().String() // ambil isi string dari pointer
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

	var val string

	// Handle pointer case
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return false // required, jadi nil tidak valid
		}
		val = field.Elem().String() // ambil isi string dari pointer
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
	birthDate, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return !birthDate.After(time.Now())
}
