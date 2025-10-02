package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// AttendanceController struct controller
type AttendanceController struct {
	attendanceService services.IAttendanceServices
	db                *gorm.DB
}

// NewAttendanceController creates a new instance of AttendanceController
func NewAttendanceController(attendanceService services.IAttendanceServices, db *gorm.DB) *AttendanceController {
	return &AttendanceController{
		attendanceService: attendanceService,
		db:                db,
	}
}

// GetAllAttendances retrieves a list of attendance records with pagination and filtering
func (ctrl *AttendanceController) GetAllAttendances(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)

	attendances, total, err := ctrl.attendanceService.GetAllFilteredAttendances(ctx, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch attendances", err.Error())
	}

	var attendanceResponses []dto.AttendanceResponse
	if copyErr := copier.Copy(&attendanceResponses, attendances); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map attendance data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Attendances fetched", attendanceResponses, meta)
}

// GetAllAttendancesByBatchSlug retrieves a list of attendance records with pagination and filtering
func (ctrl *AttendanceController) GetAllAttendancesByBatchSlug(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)
	batchSlug := c.Params("batchSlug")

	attendances, total, err := ctrl.attendanceService.GetAllFilteredAttendancesByBatchSlug(ctx, batchSlug, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch attendances", err.Error())
	}

	var attendanceResponses []dto.UserResponse
	// var attendanceResponses []dto.AttendanceResponse
	if copyErr := copier.CopyWithOption(&attendanceResponses, &attendances, copier.Option{
		IgnoreEmpty: true, // field kosong / nil tidak akan disalin
	}); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map attendance data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Attendances fetchesssd", attendanceResponses, meta)
}

// GetAttendanceByID retrieves a single attendance record by its ID
func (ctrl *AttendanceController) GetAttendanceByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	attendanceIDParam := c.Params("attendanceID")
	attendanceID, err := uuid.Parse(attendanceIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	attendance, err := ctrl.attendanceService.GetAttendanceByID(ctx, attendanceID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Attendance not found", err.Error())
	}

	var attendanceResponse dto.AttendanceResponse
	if copyErr := copier.Copy(&attendanceResponse, attendance); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map attendance data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Attendance retrieved successfully", attendanceResponse)
}

// BulkUpsertAttendance is handling for bulk attendance
func (ctrl *AttendanceController) BulkUpsertAttendance(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	body := c.Locals("body").(*dto.BulkAttendanceRequest)

	batchID, err := uuid.Parse(c.Params("batchID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid batch ID", err.Error())
	}

	results, err := ctrl.attendanceService.BulkUpsertAttendance(ctx, user, batchID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to save attendances", err.Error())
	}

	var attendanceResponses []dto.AttendanceResponse
	if copyErr := copier.Copy(&attendanceResponses, results); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map attendances data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Attendances saved", attendanceResponses)
}
