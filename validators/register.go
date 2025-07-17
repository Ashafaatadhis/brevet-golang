package validators

import "github.com/go-playground/validator/v10"

// RegisterCustomValidators registers custom validators
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("birthdate", BirthDateValidator)
	v.RegisterValidation("role_type", RoleTypeValidator)
	v.RegisterValidation("group_type", GroupTypeValidator)
	v.RegisterValidation("day_type", DayTypeValidator)
	v.RegisterValidation("course_type", CourseTypeValidator)
	v.RegisterValidation("meeting_type", MeetingTypeValidator)
	v.RegisterValidation("payment_status_type", PaymentStatusValidator)
}
