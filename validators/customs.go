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

// CourseTypeValidator checks if course_type value is valid
func CourseTypeValidator(fl validator.FieldLevel) bool {
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

	switch models.CourseType(val) {
	case models.CourseTypeOnline, models.CourseTypeOffline:
		return true
	default:
		return false
	}
}

// DayTypeValidator checks if day value is valid
func DayTypeValidator(fl validator.FieldLevel) bool {
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

	switch models.DayType(val) {
	case models.Monday, models.Tuesday, models.Wednesday, models.Thursday,
		models.Friday, models.Saturday, models.Sunday:
		return true
	default:
		return false
	}
}

// MeetingTypeValidator checks if meeting value is valid
func MeetingTypeValidator(fl validator.FieldLevel) bool {
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

	switch models.MeetingType(val) {
	case models.BasicMeeting, models.ExamMeeting:
		return true
	default:
		return false
	}
}

// AssignmentTypeValidator checks if assignment value is valid
func AssignmentTypeValidator(fl validator.FieldLevel) bool {
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

	switch models.AssignmentType(val) {
	case models.File, models.Essay:
		return true
	default:
		return false
	}
}

// AttendanceStatusValidator checks if attendance value is valid
func AttendanceStatusValidator(fl validator.FieldLevel) bool {
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

	switch models.AttendanceStatus(val) {
	case models.Present, models.Absent, models.Late:
		return true
	default:
		return false
	}
}

// PaymentStatusValidator checks if payment status value is valid
func PaymentStatusValidator(fl validator.FieldLevel) bool {
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

	switch models.PaymentStatus(val) {
	case models.Cancelled, models.Expired, models.Paid, models.Pending, models.Rejected, models.WaitingConfirmation:
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
