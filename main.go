package main

import (
	"brevet-api/config"
	"brevet-api/middlewares"
	"brevet-api/routes"
	"brevet-api/scheduler"
	"brevet-api/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return utils.ErrorResponse(c, code, "Internal Server Error", err.Error())
		},
	})

	// Middleware: logger
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middlewares.LogMiddleware())
	// Middleware: CORS dari .env
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.GetEnv("ALLOWED_ORIGINS", "*"),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Custom-Header",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowCredentials: true,
	}))

	// Static folder dari .env
	app.Static("/uploads", config.GetEnv("UPLOAD_DIR", "./public/uploads"))

	// Connect DB
	db := config.ConnectDB()

	// Initialize Redis
	config.InitRedis()

	// Cleanup expired sessions every hour
	scheduler.StartCleanupScheduler(db)
	scheduler.StartAutoSubmitScheduler(db)

	// Purpose route for testing
	app.Get("/hello", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, fiber.StatusOK, "Hello, World v2 WITH CDN!", nil)
	})

	// Register routes
	routes.RegisterRoutes(app, db)

	// Fallback 404
	app.Use(func(c *fiber.Ctx) error {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Route not found", nil)
	})

	// Start
	port := config.GetEnv("APP_PORT", "3000")
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
