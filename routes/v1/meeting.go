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

// RegisterMeetingRoutes registers all meeting-related routes
func RegisterMeetingRoutes(r fiber.Router, db *gorm.DB) {

	fileService := services.NewFileService()
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	userRepository := repository.NewUserRepository(db)
	batchRepository := repository.NewBatchRepository(db)

	assignmentRepo := repository.NewAssignmentRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	meetingRepo := repository.NewMeetingRepository(db)
	purchaseRepo := repository.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(purchaseRepo, userRepository, batchRepository, emailService, db)
	meetingService := services.NewMeetingService(meetingRepo, batchRepository, purchaseRepo, userRepository, db)
	meetingController := controllers.NewMeetingController(meetingService, db)

	assignmentRepository := repository.NewAssignmentRepository(db)
	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepo, purchaseRepo, fileService, db)
	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	materialRepository := repository.NewMaterialRepository(db)
	materialService := services.NewMaterialService(materialRepository, meetingRepo, purchaseRepo, fileService, db)
	materialController := controllers.NewMaterialController(materialService, db)

	quizRepository := repository.NewQuizRepository(db)
	quizService := services.NewQuizService(quizRepository, batchRepository, meetingRepo, attendanceRepo, assignmentRepo, submissionRepo, purchaseService, fileService, db)
	quizController := controllers.NewQuizController(quizService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), meetingController.GetAllMeetings)
	r.Get("/:id", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin", "guru"}), meetingController.GetMeetingByID)

	r.Patch("/:id", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateMeetingRequest](),
		meetingController.UpdateMeeting)
	r.Delete("/:id", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		meetingController.DeleteMeeting)

	r.Get("/:meetingID/teachers", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}), meetingController.GetTeachersByMeetingIDFiltered)

	r.Post("/:meetingID/teachers",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.AssignTeachersRequest](),
		meetingController.AddTeachersToMeeting,
	)
	r.Put("/:meetingID/teachers",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.AssignTeachersRequest](),
		meetingController.UpdateTeachersToMeeting,
	)
	r.Delete("/:meetingID/teachers/:teacherID",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		meetingController.DeleteTeachersToMeeting,
	)

	// ==================================
	// 				Assignment
	// ==================================
	r.Get("/:meetingID/assignments", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), assignmentController.GetAllAssignmentByMeetingID)
	r.Post("/:meetingID/assignments", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		middlewares.ValidateBody[dto.CreateAssignmentRequest](), assignmentController.CreateAssignment)

	// ==================================
	// 				Material
	// ==================================
	r.Get("/:meetingID/materials", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}), materialController.GetAllMaterialByMeetingID)
	r.Post("/:meetingID/materials", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		middlewares.ValidateBody[dto.CreateMaterialRequest](), materialController.CreateMaterial)
	// ==================================
	// 				Quizzes
	// ==================================
	r.Get("/:meetingID/quizzes", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin", "guru"}),
		quizController.GetQuizByMeetingIDFiltered)

	r.Post("/:meetingID/quizzes", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		middlewares.ValidateBody[dto.ImportQuizzesRequest](), quizController.CreateQuizMetadata)

}
