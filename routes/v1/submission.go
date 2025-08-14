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

// RegisterSubmissionRoutes registers all submission routes
func RegisterSubmissionRoutes(r fiber.Router, db *gorm.DB) {
	purchaseRepo := repository.NewPurchaseRepository(db)
	userRepo := repository.NewUserRepository(db)
	submissionRepository := repository.NewSubmissionRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	meetingRepository := repository.NewMeetingRepository(db)
	fileService := services.NewFileService()
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}
	batchRepository := repository.NewBatchRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, userRepo, batchRepository, emailService, db)
	submissionService := services.NewSubmissionService(submissionRepository, assignmentRepository, meetingRepository, purchaseService, fileService, db)
	submissionController := controllers.NewSubmissionController(submissionService, db)
	r.Get("/:submissionID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa", "guru"}), submissionController.GetDetailSubmission)
	r.Patch("/:submissionID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), middlewares.ValidateBody[dto.UpdateSubmissionRequest](),
		submissionController.UpdateSubmission)
	r.Delete("/:submissionID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		submissionController.DeleteSubmission)

	r.Get("/:submissionID/grade", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa", "guru"}), submissionController.GetSubmissionGrade)
	r.Put("/:submissionID/grade", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"guru"}), middlewares.ValidateBody[dto.GradeSubmissionRequest](),
		submissionController.GradeSubmission)

}
