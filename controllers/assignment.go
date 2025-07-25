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

// AssignmentController handles purchase-related operations
type AssignmentController struct {
	assignmentService *services.AssignmentService
	db                *gorm.DB
}

// NewAssignmentController creates a new instance of AssignmentController
func NewAssignmentController(assignmentService *services.AssignmentService, db *gorm.DB) *AssignmentController {
	return &AssignmentController{
		assignmentService: assignmentService,
		db:                db,
	}
}

// GetAllAssignments retrieves a list of assignments with pagination and filtering options
func (ctrl *AssignmentController) GetAllAssignments(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	assignments, total, err := ctrl.assignmentService.GetAllFilteredAssignments(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch purchases", err.Error())
	}

	var assignmentsResponse []dto.AssignmentResponse
	if copyErr := copier.Copy(&assignmentsResponse, assignments); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map assignment data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Assignments fetched", assignmentsResponse, meta)
}

// CreateAssignment creates a new assignment with the provided details
func (ctrl *AssignmentController) CreateAssignment(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateAssignmentRequest)
	user := c.Locals("user").(*utils.Claims)

	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	assignment, err := ctrl.assignmentService.CreateAssignment(user, meetingID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to create assignment", err.Error())
	}

	var assignmentResponse dto.AssignmentResponse
	if copyErr := copier.Copy(&assignmentResponse, assignment); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map assignment data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Assignment created successfully", assignmentResponse)
}
