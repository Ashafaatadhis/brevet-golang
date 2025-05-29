package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse returns a success response
func SuccessResponse(c *fiber.Ctx, statusCode int, msg string, data any) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data":    data,
	})
}

// SuccessWithMeta returns a success response with meta data
func SuccessWithMeta(c *fiber.Ctx, statusCode int, msg string, data any, meta any) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": true,
		"message": msg,
		"data":    data,
		"meta":    meta,
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c *fiber.Ctx, code int, msg string, err any) error {
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": msg,
		"error":   err,
	})
}
