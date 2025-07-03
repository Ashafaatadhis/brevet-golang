package validators

import "github.com/go-playground/validator/v10"

// RegisterCustomValidators registers custom validators
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("birthdate", BirthDateValidator)
	v.RegisterValidation("role_type", RoleTypeValidator)
	v.RegisterValidation("group_type", GroupTypeValidator)
	v.RegisterValidation("course_type", CourseTypeValidator)
}
