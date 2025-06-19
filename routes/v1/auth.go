package v1

import (
	"brevet-api/controllers"
	"brevet-api/dto"
	"brevet-api/middlewares"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	authService := services.NewAuthService(db)

	verificationService := services.NewVerificationService(db)
	authController := controllers.NewAuthController(authService, verificationService, db)

	r.Post("/register", middlewares.ValidateBody[dto.RegisterRequest](), authController.Register)
	r.Post("/login", middlewares.ValidateBody[dto.LoginRequest](), authController.Login)
	r.Post("/verify", middlewares.ValidateBody[dto.VerifyRequest](), authController.VerifyCode) // Add this line
	r.Post("/resend-verification", middlewares.ValidateBody[dto.ResendVerificationRequest](), authController.ResendVerification)
	r.Post("/refresh-token", authController.RefreshToken)
	r.Delete("/logout", middlewares.RequireAuth(), authController.Logout)

}
