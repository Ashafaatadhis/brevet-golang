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

	purchaseRepo := repository.NewPurchaseRepository(db)

	meetingRepo := repository.NewMeetingRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepo, purchaseRepo, fileService, db)

	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), assignmentController.GetAllAssignments)
	r.Get("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru", "siswa"}), assignmentController.GetAssignmentByID)
	r.Patch("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), middlewares.ValidateBody[dto.UpdateAssignmentRequest](),
		assignmentController.UpdateAssignment)
	r.Delete("/:assignmentID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), assignmentController.DeleteAssignment)

	// ==================================
	// 				Submissions
	// ==================================
	submissionRepository := repository.NewSubmissionRepository(db)
	meetingRepository := repository.NewMeetingRepository(db)
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}
	batchRepository := repository.NewBatchRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, batchRepository, emailService, db)
	submissionService := services.NewSubmissionService(submissionRepository, assignmentRepository, meetingRepository, purchaseService, fileService, db)
	submissionController := controllers.NewSubmissionController(submissionService, db)
	r.Get("/:assignmentID/submissions", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa", "guru"}), submissionController.GetAllSubmissionByAssignmentID)

	r.Post("/:assignmentID/submissions", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), middlewares.ValidateBody[dto.CreateSubmissionRequest](),
		submissionController.CreateSubmission)

}
