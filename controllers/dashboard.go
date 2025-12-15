package controllers

import (
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DashboardController handles dashboard-related requests
type DashboardController struct {
	service services.IDashboardService
	db      *gorm.DB
}

// NewDashboardController creates a new dashboard controller
func NewDashboardController(service services.IDashboardService, db *gorm.DB) *DashboardController {
	return &DashboardController{
		service: service,
		db:      db,
	}
}

// GetAdminDashboard returns admin dashboard statistics
func (c *DashboardController) GetAdminDashboard(ctx *fiber.Ctx) error {

	period := ctx.Query("period", "30d")

	if period != "7d" && period != "30d" && period != "90d" {
		return utils.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid period. Must be 7d, 30d, or 90d", nil)
	}

	// Get dashboard data
	dashboard, err := c.service.GetAdminDashboard(ctx.Context(), period)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Dashboard data retrieved successfully", dashboard)
}

// GetTeacherDashboard returns dashboard statistics for a teacher
func (c *DashboardController) GetTeacherDashboard(ctx *fiber.Ctx) error {
	userRaw := ctx.Locals("user")
	claims, ok := userRaw.(*utils.Claims)
	if !ok || claims == nil {
		return utils.ErrorResponse(ctx, fiber.StatusUnauthorized, "Unauthorized: invalid user data", nil)
	}

	dashboard, err := c.service.GetTeacherDashboard(ctx.Context(), claims.UserID)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Dashboard data retrieved successfully", dashboard)
}

// GetStudentScoreProgress returns score progress per meeting for a student in a batch
func (c *DashboardController) GetStudentScoreProgress(ctx *fiber.Ctx) error {
	userRaw := ctx.Locals("user")
	claims, ok := userRaw.(*utils.Claims)
	if !ok || claims == nil {
		return utils.ErrorResponse(ctx, fiber.StatusUnauthorized, "Unauthorized: invalid user data", nil)
	}

	batchSlug := ctx.Query("batch_slug", "")
	if batchSlug == "" {
		return utils.ErrorResponse(ctx, fiber.StatusBadRequest, "batch_slug is required", nil)
	}

	data, err := c.service.GetStudentScoreProgress(ctx.Context(), batchSlug, claims.UserID)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Score progress retrieved successfully", data)
}

// GetStudentDashboard returns dashboard stats for a student
func (c *DashboardController) GetStudentDashboard(ctx *fiber.Ctx) error {
	userRaw := ctx.Locals("user")
	claims, ok := userRaw.(*utils.Claims)
	if !ok || claims == nil {
		return utils.ErrorResponse(ctx, fiber.StatusUnauthorized, "Unauthorized: invalid user data", nil)
	}

	data, err := c.service.GetStudentDashboard(ctx.Context(), claims.UserID)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Dashboard data retrieved successfully", data)
}

// GetRevenueChart returns revenue chart data per day
func (c *DashboardController) GetRevenueChart(ctx *fiber.Ctx) error {
	// Default period adalah 30d
	period := ctx.Query("period", "30d")

	// Validasi period
	if period != "7d" && period != "30d" && period != "90d" {
		return utils.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid period. Must be 7d, 30d, or 90d", nil)
	}

	// Get revenue chart data
	chartData, err := c.service.GetRevenueChart(ctx.Context(), period)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Revenue chart data retrieved successfully", chartData)
}

// GetPendingPayments returns list of pending payments that need verification
func (c *DashboardController) GetPendingPayments(ctx *fiber.Ctx) error {
	limit := ctx.QueryInt("limit", 10)

	data, err := c.service.GetPendingPayments(ctx.Context(), limit)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Pending payments retrieved successfully", data)
}

// GetBatchProgress returns list of batches with progress
func (c *DashboardController) GetBatchProgress(ctx *fiber.Ctx) error {
	limit := ctx.QueryInt("limit", 10)

	data, err := c.service.GetBatchProgress(ctx.Context(), limit)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Batch progress retrieved successfully", data)
}

// GetTeacherWorkload returns teacher workload statistics
func (c *DashboardController) GetTeacherWorkload(ctx *fiber.Ctx) error {
	period := ctx.Query("period", "week") // week or month

	data, err := c.service.GetTeacherWorkload(ctx.Context(), period)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Teacher workload retrieved successfully", data)
}

// GetCertificateStats returns certificate statistics
func (c *DashboardController) GetCertificateStats(ctx *fiber.Ctx) error {
	period := ctx.Query("period", "30d") // 7d, 30d, or 90d

	if period != "7d" && period != "30d" && period != "90d" {
		return utils.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid period. Must be 7d, 30d, or 90d", nil)
	}

	data, err := c.service.GetCertificateStats(ctx.Context(), period)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Certificate stats retrieved successfully", data)
}

// GetRecentActivities returns recent activities in the system
func (c *DashboardController) GetRecentActivities(ctx *fiber.Ctx) error {
	period := ctx.Query("period", "30d")
	limit := ctx.QueryInt("limit", 20)

	if period != "7d" && period != "30d" && period != "90d" {
		return utils.ErrorResponse(ctx, fiber.StatusBadRequest, "Invalid period. Must be 7d, 30d, or 90d", nil)
	}

	data, err := c.service.GetRecentActivities(ctx.Context(), period, limit)
	if err != nil {
		return utils.ErrorResponse(ctx, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return utils.SuccessResponse(ctx, fiber.StatusOK, "Recent activities retrieved successfully", data)
}
