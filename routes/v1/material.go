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

// RegisterMaterialRoutes register material routes
func RegisterMaterialRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller

	fileService := services.NewFileService()

	purchaseRepo := repository.NewPurchaseRepository(db)

	meetingRepo := repository.NewMeetingRepository(db)
	materialRepository := repository.NewMaterialRepository(db)
	materialService := services.NewMaterialService(materialRepository, meetingRepo, purchaseRepo, fileService, db)

	materialController := controllers.NewMaterialController(materialService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), materialController.GetAllMaterials)
	r.Get("/:materialID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru", "siswa"}), materialController.GetMaterialByID)
	r.Patch("/:materialID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), middlewares.ValidateBody[dto.UpdateMaterialRequest](),
		materialController.UpdateMaterial)
	r.Delete("/:materialID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), materialController.DeleteMaterial)

}
