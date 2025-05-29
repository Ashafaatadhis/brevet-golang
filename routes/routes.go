package routes

import (
	v1 "brevet-api/routes/v1"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api")
	v1.RegisterV1Routes(api.Group("/v1"), db)
}
