package v1

import (
	"brevet-api/controllers"
	"brevet-api/dto"
	"brevet-api/middlewares"

	"github.com/gofiber/fiber/v2"
)

// RegisterUploadRoutes registers upload routes
func RegisterUploadRoutes(r fiber.Router) {
	// Routes untuk upload
	imageGroup := r.Group("/images")
	imageGroup.Post("/", middlewares.ValidateBody[dto.UploadRequest](), controllers.UploadImage)

	docGroup := r.Group("/documents")
	docGroup.Post("/", middlewares.ValidateBody[dto.UploadRequest](), controllers.UploadDocument)

	// Hapus file (umum untuk semua jenis file)
	r.Delete("/", middlewares.ValidateBody[dto.DeleteRequest](), controllers.DeleteFile)
}
