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
	authService := services.NewAuthService(authRepository, verificationService, sessionRepository)

	userRepository := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepository, db, authRepository)
	userController := controllers.NewUserController(userService, authService, db)

	// ========================
	//          BATCH
	// ========================
	batchRepository := repository.NewBatchRepository(db)
	courseRepository := repository.NewCourseRepository(db)
	fileService := services.NewFileService()
	courseService := services.NewCourseService(courseRepository, db, fileService)
	batchService := services.NewBatchService(batchRepository, userRepository, courseRepository, db, fileService)
	meetingRepository := repository.NewMeetingRepository(db)
	purchaseRepo := repository.NewPurchaseRepository(db)
	meetingService := services.NewMeetingService(meetingRepository, batchRepository, purchaseRepo, userRepository, db)

	batchController := controllers.NewBatchController(batchService, meetingService, courseService, db)

	purchaseService := services.NewPurchaseService(purchaseRepo, batchRepository, db)
	purchaseController := controllers.NewPurchaseController(purchaseService, db)

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

	// r.Get("/batches", middlewares.RequireAuth(),
	// 	middlewares.RequireRole([]string{"guru", "siswa"}), batchController.GetMyBatchesByID)
}
