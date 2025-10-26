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

// RegisterMeRoutes registers all me-related routes
func RegisterMeRoutes(r fiber.Router, db *gorm.DB) {

	// =======================
	//         USER
	// =======================
	authRepository := repository.NewAuthRepository(db)
	sessionRepository := repository.NewUserSessionRepository(db)
	verificationRepository := repository.NewVerificationRepository(db)
	verificationService := services.NewVerificationService(verificationRepository)
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}
	tokenService := services.NewTokenService()
	authService := services.NewAuthService(authRepository, verificationService, sessionRepository, tokenService, emailService)

	userRepository := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepository, db, authRepository)
	userController := controllers.NewUserController(userService, authService, db)

	// ========================
	//          BATCH
	// ========================
	batchRepository := repository.NewBatchRepository(db)
	courseRepository := repository.NewCourseRepository(db)
	quizRepository := repository.NewQuizRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	submissionRepository := repository.NewSubmissionRepository(db)
	attendanceRepository := repository.NewAttendanceRepository(db)
	meetingRepository := repository.NewMeetingRepository(db)
	fileService := services.NewFileService()
	courseService := services.NewCourseService(courseRepository, db, fileService)

	batchService := services.NewBatchService(batchRepository, userRepository, quizRepository, courseRepository, assignmentRepository, submissionRepository, attendanceRepository, meetingRepository, db, fileService)
	purchaseRepo := repository.NewPurchaseRepository(db)
	meetingService := services.NewMeetingService(meetingRepository, batchRepository, purchaseRepo, userRepository, db)

	batchController := controllers.NewBatchController(batchService, meetingService, courseService, db)

	purchaseService := services.NewPurchaseService(purchaseRepo, userRepository, batchRepository, emailService, db)
	purchaseController := controllers.NewPurchaseController(purchaseService, db)

	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepository, purchaseRepo, fileService, db)

	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	quizService := services.NewQuizService(quizRepository, batchRepository, meetingRepository, attendanceRepository, assignmentRepository, submissionRepository, purchaseService, fileService, db)
	quizController := controllers.NewQuizController(quizService, db)

	certificateRepository := repository.NewCertificateRepository(db)
	certificateService := services.NewCertificateService(certificateRepository, userRepository, batchRepository, attendanceRepository, meetingRepository, purchaseService, batchService, fileService)
	certificateController := controllers.NewCertificateController(certificateService)

	scoreController := controllers.NewScoreController(services.NewScoreService(db, batchRepository, meetingRepository, purchaseService, quizRepository, submissionRepository), db)

	r.Get("/", middlewares.RequireAuth(), userController.GetProfile)
	r.Patch("/",
		middlewares.RequireAuth(),
		middlewares.ValidateBody[dto.UpdateMyProfile](),
		userController.UpdateMyProfile,
	)

	r.Get("/purchases", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), purchaseController.GetMyPurchase)
	r.Get("/purchases/:id", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), purchaseController.GetMyPurchaseByID)
	r.Post("/purchases", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), middlewares.ValidateBody[dto.CreatePurchase](), purchaseController.CreatePurchase)

	r.Patch("/purchases/:id/pay", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), middlewares.ValidateBody[dto.PayPurchaseRequest](), purchaseController.Pay)

	r.Patch("/purchases/:id/cancel", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), purchaseController.Cancel)

	r.Get("/batches", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"guru", "siswa"}), batchController.GetMyBatches)

	r.Get("/batches/:batchID/progress", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), batchController.GetProgress)

	r.Post("/batches/:batchID/certificate", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}), certificateController.GenerateCertificate)

	r.Get("/batches/:batchID/certificate",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		certificateController.GetCertificate)

	r.Get("/batches/:batchID/scores",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		scoreController.GetScores)

	r.Get("/assignments/upcoming",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		assignmentController.GetAllUpcomingAssignments)

	r.Get("/quizzes/upcoming",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.GetAllUpcomingQuizzes)

	// r.Get("/batches", middlewares.RequireAuth(),
	// 	middlewares.RequireRole([]string{"guru", "siswa"}), batchController.GetMyBatchesByID)
}
