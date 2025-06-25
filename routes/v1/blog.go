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

// RegisterBlogRoutes registers all blog-related routes
func RegisterBlogRoutes(r fiber.Router, db *gorm.DB) {
	blogRepository := repository.NewBlogRepository(db)
	fileService := services.NewFileService()
	blogService := services.NewBlogService(blogRepository, db, fileService)
	blogController := controllers.NewBlogController(blogService, db)

	r.Get("/", blogController.GetAllBlogs)
	r.Get("/:slug", blogController.GetBlogBySlug)
	r.Post("/",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.CreateBlogRequest](),
		blogController.CreateBlog,
	)

	r.Put("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		middlewares.ValidateBody[dto.UpdateBlogRequest](),
		blogController.UpdateBlog,
	)

	r.Delete("/:id",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		blogController.DeleteBlog,
	)
}
