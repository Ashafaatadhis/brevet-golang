package v1

import (
	"brevet-api/controllers"
	"brevet-api/middlewares"
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterPurchaseRoutes registers all purchase-related routes
func RegisterPurchaseRoutes(r fiber.Router, db *gorm.DB) {

	purchaseRepo := repository.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, db)
	purchaseController := controllers.NewPurchaseController(purchaseService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), purchaseController.GetAllPurchases)

	r.Get("/:id", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), purchaseController.GetPurchaseByID)

}
