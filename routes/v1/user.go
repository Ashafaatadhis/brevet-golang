package v1

import (
	"brevet-api/controllers"
	"brevet-api/dto"
	"brevet-api/middlewares" // Import your middleware package
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	authService := services.NewAuthService(db)

	userService := services.NewUserService(db)
	userController := controllers.NewUserController(userService, authService, db)

	// Protected route example (requires authentication)
	r.Get("/me", middlewares.RequireAuth(), userController.GetProfile)

	// Public route
	r.Get("/", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}), userController.GetAllUsers)
	r.Get("/:id", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}), userController.GetUserByID)
	r.Post("/", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.CreateUserWithProfileRequest](), userController.CreateUserWithProfile)
	r.Put("/:id", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateUserWithProfileRequest](), userController.UpdateUserWithProfile)
	r.Delete("/:id", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}),
		userController.DeleteUserByID)

}
