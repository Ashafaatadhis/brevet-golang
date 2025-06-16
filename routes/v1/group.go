package v1

import (
	"brevet-api/controllers"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterGroupRoutes registers group-related routes
func RegisterGroupRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	groupService := services.NewGroupService(db)
	groupController := controllers.NewGroupController(groupService, db)

	// Routes untuk grup
	r.Get("/", groupController.GetAllGroups)    // Get all groups
	r.Get("/:id", groupController.GetGroupByID) // Get group by ID                                      // Delete a group
}
