package middlewares

import (
	"brevet-api/utils"
	"brevet-api/validators" // import path sesuai folder kamu

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validators.RegisterCustomValidators(validate)
}

// ValidateBody is a middleware that validates the body of the request
func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T

		if err := c.BodyParser(&body); err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Gagal parsing body", err.Error())
		}

		if err := validate.Struct(body); err != nil {
			if ve, ok := err.(validator.ValidationErrors); ok {
				errMap := validators.FormatValidationError(ve)
				return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validasi gagal", errMap)
			}
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Validasi gagal", err.Error())
		}

		c.Locals("body", &body)
		return c.Next()
	}
}
