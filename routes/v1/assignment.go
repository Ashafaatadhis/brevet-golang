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

// RegisterAssignmentRoutes register assignment routes
func RegisterAssignmentRoutes(r fiber.Router, db *gorm.DB) {
	// Inisialisasi service dan controller

	fileService := services.NewFileService()

	meetingRepo := repository.NewMeetingRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepo, fileService, db)

	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), assignmentController.GetAllAssignments)
	r.Get("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), assignmentController.GetAssignmentByID)
	r.Patch("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), middlewares.ValidateBody[dto.UpdateAssignmentRequest](),
		assignmentController.UpdateAssignment)
	r.Delete("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), assignmentController.DeleteAssignment)

}
