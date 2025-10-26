package v1

import (
	"brevet-api/controllers"
	"brevet-api/middlewares"
	"brevet-api/repository"
	"brevet-api/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterDashboardRoutes registers all dashboard-related routes
func RegisterDashboardRoutes(r fiber.Router, db *gorm.DB) {
	purchaseRepository := repository.NewPurchaseRepository(db)
	batchRepository := repository.NewBatchRepository(db)
	certificateRepository := repository.NewCertificateRepository(db)

	dashboardService := services.NewDashboardService(purchaseRepository, batchRepository, certificateRepository, db)
	dashboardController := controllers.NewDashboardController(dashboardService, db)

	// Main dashboard stats
	r.Get("/admin",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetAdminDashboard,
	)

	// Revenue chart
	r.Get("/admin/revenue-chart",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetRevenueChart,
	)

	// Pending payments
	r.Get("/admin/pending-payments",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetPendingPayments,
	)

	// Batch progress
	r.Get("/admin/batch-progress",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetBatchProgress,
	)

	// Teacher workload
	r.Get("/admin/teacher-workload",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetTeacherWorkload,
	)

	// Certificate stats
	r.Get("/admin/certificate-stats",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetCertificateStats,
	)

	// Recent activities
	r.Get("/admin/recent-activities",
		middlewares.RequireAuth(),
		middlewares.RequireRole([]string{"admin"}),
		dashboardController.GetRecentActivities,
	)
}
