package v1

import (
	"brevet-api/controllers"
	"brevet-api/middlewares"
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterAssignmentRoutes register assignment routes
func RegisterAssignmentRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller
	meetingRepo := repository.NewMeetingRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepo, db)

	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), assignmentController.GetAllAssignments)

}
