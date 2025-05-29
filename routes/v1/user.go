package v1

import (
	"brevet-api/controllers"
	"brevet-api/middlewares" // Import your middleware package
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	userService := services.NewUserService(db)
	userController := controllers.NewUserController(userService)

	// Public route
	r.Get("/", userController.GetAllUsers)

	// Protected route example (requires authentication)
	r.Get("/me", middlewares.RequireAuth(), userController.GetProfile)
}
