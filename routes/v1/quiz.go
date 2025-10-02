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

// RegisterQuizRoutes registers all quiz-related routes
func RegisterQuizRoutes(r fiber.Router, db *gorm.DB) {
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	userRepository := repository.NewUserRepository(db)
	fileService := services.NewFileService()
	batchRepository := repository.NewBatchRepository(db)
	meetingRepo := repository.NewMeetingRepository(db)

	assignmentRepo := repository.NewAssignmentRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	purchaseRepo := repository.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, userRepository, batchRepository, emailService, db)

	quizRepository := repository.NewQuizRepository(db)
	quizService := services.NewQuizService(quizRepository, batchRepository, meetingRepo, attendanceRepo, assignmentRepo, submissionRepo, purchaseService, fileService, db)
	quizController := controllers.NewQuizController(quizService, db)

	r.Post("/attempts/:attemptID/temp-submissions",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		middlewares.ValidateBody[dto.SaveTempSubmissionRequest](),
		quizController.SaveTempSubmission,
	)

	r.Post("/attempts/:attemptID/submissions",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.SubmitQuiz,
	)

	r.Get("/attempts/:attemptID",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.GetAttemptDetail,
	)

	r.Get("/attempts/:attemptID/result",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa", "guru", "admin"}),
		quizController.GetAttemptResult,
	)

	r.Get("/:quizID", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin", "guru", "siswa"}),
		quizController.GetQuizByID)

	r.Get("/:quizID/questions", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin", "guru"}),
		quizController.GetQuizWithQuestions)

	r.Get("/:quizID/attempts/active",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.GetActiveAttempt,
	)
	r.Get("/:quizID/attempts",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.GetListAttempt,
	)

	r.Post("/:quizID/import-questions",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		quizController.ImportQuestionsFromExcel)

	r.Post("/:quizID/start", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		quizController.StartQuiz)

	r.Patch("/:quizID",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		middlewares.ValidateBody[dto.UpdateQuizRequest](),
		quizController.UpdateQuiz,
	)

	r.Delete("/:quizID",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		quizController.DeleteQuiz,
	)

}
