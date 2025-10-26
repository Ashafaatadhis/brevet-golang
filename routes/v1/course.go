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

// RegisterCourseRoutes registers all course-related routes
func RegisterCourseRoutes(r fiber.Router, db *gorm.DB) {
	courseRepository := repository.NewCourseRepository(db)
	fileService := services.NewFileService()
	courseService := services.NewCourseService(courseRepository, db, fileService)
	courseController := controllers.NewCourseController(courseService, db)

	batchRepository := repository.NewBatchRepository(db)
	userRepository := repository.NewUserRepository(db)
	quizRepository := repository.NewQuizRepository(db)
	assignmentRepository := repository.NewAssignmentRepository(db)
	submissionRepository := repository.NewSubmissionRepository(db)
	attendanceRepository := repository.NewAttendanceRepository(db)
	meetingRepository := repository.NewMeetingRepository(db)
	batchService := services.NewBatchService(batchRepository, userRepository, quizRepository, courseRepository, assignmentRepository, submissionRepository, attendanceRepository, meetingRepository, db, fileService)
	purchaseRepo := repository.NewPurchaseRepository(db)
	meetingService := services.NewMeetingService(meetingRepository, batchRepository, purchaseRepo, userRepository, db)

	batchController := controllers.NewBatchController(batchService, meetingService, courseService, db)

	r.Get("/", courseController.GetAllCourses)
	r.Get("/:slug", courseController.GetCourseBySlug)
	r.Post("/",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.CreateCourseRequest](),
		courseController.CreateCourse,
	)
	r.Put("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateCourseRequest](),
		courseController.UpdateCourse,
	)
	r.Delete("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		courseController.DeleteCourse,
	)

	r.Get("/:courseSlug/batches", batchController.GetBatchByCourseSlug)

	r.Post("/:courseId/batches",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.CreateBatchRequest](),
		batchController.CreateBatch,
	)

}
