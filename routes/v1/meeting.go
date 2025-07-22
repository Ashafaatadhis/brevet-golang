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

	userRepository := repository.NewUserRepository(db)
	batchRepository := repository.NewBatchRepository(db)

	meetingRepo := repository.NewMeetingRepository(db)
	meetingService := services.NewMeetingService(meetingRepo, batchRepository, userRepository, db)
	meetingController := controllers.NewMeetingController(meetingService, db)

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

}
