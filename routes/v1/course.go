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
	fileServce := services.NewFileService()
	courseService := services.NewCourseService(courseRepository, db, fileServce)
	courseController := controllers.NewCourseController(courseService, db)

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
}
