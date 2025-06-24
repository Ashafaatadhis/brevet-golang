package v1

import (
	"brevet-api/controllers"
	"brevet-api/dto"
	"brevet-api/middlewares" // Import your middleware package
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(r fiber.Router, db *gorm.DB) {

	authRepository := repository.NewAuthRepository(db)
	sessionRepository := repository.NewUserSessionRepository(db)
	verificationRepository := repository.NewVerificationRepository(db)
	verificationService := services.NewVerificationService(verificationRepository)
	authService := services.NewAuthService(authRepository, verificationService, sessionRepository)

	userRepository := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepository, db, authRepository)
	userController := controllers.NewUserController(userService, authService, db)

	r.Get("/me", middlewares.RequireAuth(), userController.GetProfile)
	r.Put("/me",
		middlewares.RequireAuth(),
		middlewares.ValidateBody[dto.UpdateMyProfile](),
		userController.UpdateMyProfile,
	)
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
