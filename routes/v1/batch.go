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

// RegisterBatchRoute registers all batch-related routes
func RegisterBatchRoute(r fiber.Router, db *gorm.DB) {

	batchRepository := repository.NewBatchRepository(db)
	userRepository := repository.NewUserRepository(db)
	courseRepository := repository.NewCourseRepository(db)
	fileService := services.NewFileService()
	courseService := services.NewCourseService(courseRepository, db, fileService)
	batchService := services.NewBatchService(batchRepository, userRepository, courseRepository, db, fileService)

	meetingRepository := repository.NewMeetingRepository(db)
	meetingService := services.NewMeetingService(meetingRepository, batchRepository, userRepository, db)

	batchController := controllers.NewBatchController(batchService, meetingService, courseService, db)

	meetingController := controllers.NewMeetingController(meetingService, db)

	r.Get("/", batchController.GetAllBatches)
	r.Get("/:slug", batchController.GetBatchBySlug)
	// POST /v1/courses/:courseId/batches
	r.Put("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateBatchRequest](),
		batchController.UpdateBatch,
	)
	r.Delete("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		batchController.DeleteBatch,
	)

	r.Get("/:batchSlug/meetings", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "siswa", "guru"}),
		meetingController.GetMeetingsByBatchSlug)
	r.Post("/:batchID/meetings", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.CreateMeetingRequest](),
		meetingController.CreateMeeting)

	// Get All Students
	r.Get("/:batchSlug/students", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		batchController.GetAllStudents)

	// THIS IS ROUTE FOR ASSIGN TEACHER TO BATCH
	// 	Method	Route	Deskripsi
	// POST	/batches/:batchID/teachers	Tambah teacher ke batch tertentu
	// GET	/batches/:batchID/teachers	List semua teacher dalam satu batch
	// DELETE	/batches/:batchID/teachers/:userID	Hapus teacher tertentu dari

}
