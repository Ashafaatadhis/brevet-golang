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

	userRepository := repository.NewUserRepository(db)
	batchRepository := repository.NewBatchRepository(db)

	meetingRepo := repository.NewMeetingRepository(db)
	purchaseRepo := repository.NewPurchaseRepository(db)
	meetingService := services.NewMeetingService(meetingRepo, batchRepository, purchaseRepo, userRepository, db)
	meetingController := controllers.NewMeetingController(meetingService, db)

	assignmentRepository := repository.NewAssignmentRepository(db)
	assignmentService := services.NewAssignmentService(assignmentRepository, meetingRepo, purchaseRepo, fileService, db)
	assignmentController := controllers.NewAssignmentController(assignmentService, db)

	materialRepository := repository.NewMaterialRepository(db)
	materialService := services.NewMaterialService(materialRepository, meetingRepo, purchaseRepo, fileService, db)
	materialController := controllers.NewMaterialController(materialService, db)

	meetingRepository := repository.NewMeetingRepository(db)
	purchaseRepository := repository.NewPurchaseRepository(db)
	attendanceRepository := repository.NewAttendanceRepository(db)
	attendanceService := services.NewAttendanceService(attendanceRepository, meetingRepository, purchaseRepository, db)
	attendanceController := controllers.NewAttendanceController(attendanceService, db)

	r.Get("/", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}), meetingController.GetAllMeetings)
	r.Get("/:id", middlewares.RequireAuth(), middlewares.RequireRole([]string{"admin"}), meetingController.GetMeetingByID)

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
		middlewares.RequireRole([]string{"admin", "guru"}), materialController.GetAllMaterials)
	r.Post("/:meetingID/materials", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin", "guru"}),
		middlewares.ValidateBody[dto.CreateMaterialRequest](), materialController.CreateMaterial)

	// ==================================
	// 				Attendance
	// ==================================
	r.Put("/:meetingID/attendances/bulk", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.BulkAttendanceRequest](), attendanceController.BulkUpsertAttendance)

}
