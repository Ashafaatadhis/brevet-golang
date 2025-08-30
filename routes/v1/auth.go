package v1

import (
	"brevet-api/controllers"
	"brevet-api/dto"
	"brevet-api/middlewares"
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	authRepository := repository.NewAuthRepository(db)
	sessionRepository := repository.NewUserSessionRepository(db)
	verificationRepository := repository.NewVerificationRepository(db) // Assuming you have a verification repository
	verificationService := services.NewVerificationService(verificationRepository)
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}
	tokenService := services.NewTokenService()
	authService := services.NewAuthService(authRepository, verificationService, sessionRepository, tokenService, emailService)

	authController := controllers.NewAuthController(authService, verificationService, db)

	r.Post("/register", middlewares.ValidateBody[dto.RegisterRequest](), authController.Register)
	r.Post("/login", middlewares.ValidateBody[dto.LoginRequest](), authController.Login)
	r.Post("/verify", middlewares.ValidateBody[dto.VerifyRequest](), authController.VerifyCode) // Add this line
	r.Post("/resend-verification", middlewares.ValidateBody[dto.ResendVerificationRequest](), authController.ResendVerification)
	r.Post("/refresh-token", authController.RefreshToken)
	r.Delete("/logout", middlewares.RequireAuth(), authController.Logout)

}
