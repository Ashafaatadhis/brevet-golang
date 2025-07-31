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

// GetAssignmentByID retrieves a single assignment by its ID
func (ctrl *AssignmentController) GetAssignmentByID(c *fiber.Ctx) error {
	assignmentIDParam := c.Params("assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	assignment, err := ctrl.assignmentService.GetAssignmentByID(assignmentID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Assignment not found", err.Error())
	}

	var assignmentResponse dto.AssignmentResponse
	if copyErr := copier.Copy(&assignmentResponse, assignment); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map assignment data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Assignment retrieved successfully", assignmentResponse)
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

// UpdateAssignment updates an existing assignment and its files
func (ctrl *AssignmentController) UpdateAssignment(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.UpdateAssignmentRequest)
	user := c.Locals("user").(*utils.Claims)

	assignmentIDParam := c.Params("assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	assignment, err := ctrl.assignmentService.UpdateAssignment(user, assignmentID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update assignment", err.Error())
	}

	var assignmentResponse dto.AssignmentResponse
	if copyErr := copier.Copy(&assignmentResponse, assignment); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map assignment data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Assignment updated successfully", assignmentResponse)
}

// DeleteAssignment deletes an existing assignment and its related files
func (ctrl *AssignmentController) DeleteAssignment(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.Claims)

	assignmentIDParam := c.Params("assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.assignmentService.DeleteAssignment(user, assignmentID); err != nil {
		return utils.ErrorResponse(c, 400, "Failed to delete assignment", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Assignment deleted successfully", nil)
}
