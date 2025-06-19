package v1

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterV1Routes registers all v1 API routes
func RegisterV1Routes(r fiber.Router, db *gorm.DB) {
	// /v1/auth
	authGroup := r.Group("/auth")
	RegisterAuthRoutes(authGroup, db)

	// /v1/uploads
	uploadGroup := r.Group("/upload")
	RegisterUploadRoutes(uploadGroup)

	// /v1/users
	userGroup := r.Group("/users")
	RegisterUserRoutes(userGroup, db)

}
