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

// RegisterTestimonialRoute registers all testimonial routes
func RegisterTestimonialRoute(r fiber.Router, db *gorm.DB) {

	testimonialRepository := repository.NewTestimonialRepository(db)
	userRepository := repository.NewUserRepository(db)
	purchaseRepository := repository.NewPurchaseRepository(db)
	batchRepository := repository.NewBatchRepository(db)
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	purchaseService := services.NewPurchaseService(purchaseRepository, userRepository, batchRepository, emailService, db)
	testimonialService := services.NewTestimonialService(testimonialRepository, purchaseService, batchRepository)

	testimonialController := controllers.NewTestimonialController(testimonialService)

	r.Get("/", testimonialController.GetAllFiltered)
	r.Get("/:testimonialID", testimonialController.GetByID)
	r.Patch("/:testimonialID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		middlewares.ValidateBody[dto.UpdateTestimonialRequest](),
		testimonialController.Update)
	r.Delete("/:testimonialID", middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"siswa"}),
		testimonialController.Delete)
}
