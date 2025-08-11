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

// RegisterPurchaseRoutes registers all purchase-related routes
func RegisterPurchaseRoutes(r fiber.Router, db *gorm.DB) {

	purchaseRepo := repository.NewPurchaseRepository(db)
	batchRepo := repository.NewBatchRepository(db)
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	purchaseService := services.NewPurchaseService(purchaseRepo, batchRepo, emailService, db)
	purchaseController := controllers.NewPurchaseController(purchaseService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), purchaseController.GetAllPurchases)

	r.Get("/:id", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), purchaseController.GetPurchaseByID)
	r.Patch("/:id/status", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateStatusPayment](),
		purchaseController.UpdateStatusPayment)

}
