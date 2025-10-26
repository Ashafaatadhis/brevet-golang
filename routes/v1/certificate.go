package v1

import (
	"brevet-api/controllers"
	"brevet-api/middlewares"
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterCertificateRoutes registers all me-related routes
func RegisterCertificateRoutes(r fiber.Router, db *gorm.DB) {

	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	userRepository := repository.NewUserRepository(db)

	batchRepository := repository.NewBatchRepository(db)
	courseRepository := repository.NewCourseRepository(db)
	quizRepository := repository.NewQuizRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	submissionRepository := repository.NewSubmissionRepository(db)
	attendanceRepository := repository.NewAttendanceRepository(db)
	meetingRepository := repository.NewMeetingRepository(db)
	fileService := services.NewFileService()

	batchService := services.NewBatchService(batchRepository, userRepository, quizRepository, courseRepository, assignmentRepository, submissionRepository, attendanceRepository, meetingRepository, db, fileService)
	purchaseRepo := repository.NewPurchaseRepository(db)

	purchaseService := services.NewPurchaseService(purchaseRepo, userRepository, batchRepository, emailService, db)

	certificateRepository := repository.NewCertificateRepository(db)
	certificateService := services.NewCertificateService(certificateRepository, userRepository, batchRepository, attendanceRepository, meetingRepository, purchaseService, batchService, fileService)
	certificateController := controllers.NewCertificateController(certificateService)
	r.Get("/number/:number", certificateController.GetByNumber)
	r.Get("/:certificateID",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru", "siswa"}),
		certificateController.GetBatchCertificate)

	// routes
	r.Get("/:certificateID/verify", certificateController.VerifyCertificate)

}
