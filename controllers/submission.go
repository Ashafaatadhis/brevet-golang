package controllers

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// SubmissionController is struct
type SubmissionController struct {
	submissionService services.ISubmissionService
	db                *gorm.DB
}

// NewSubmissionController creates a new instance of SubmissionController
func NewSubmissionController(submissionService services.ISubmissionService, db *gorm.DB) *SubmissionController {
	return &SubmissionController{
		submissionService: submissionService,
		db:                db,
	}
}

// GetAllSubmissionByAssignmentID for get all
func (ctrl *SubmissionController) GetAllSubmissionByAssignmentID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	assignmentID, err := uuid.Parse(c.Params("assignmentID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid assignment ID", err.Error())
	}

	opts := utils.ParseQueryOptions(c)

	submissions, total, err := ctrl.submissionService.GetAllSubmissionsByAssignmentUser(ctx, assignmentID, user, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch submissions", err.Error())
	}

	var submissionsResponse []dto.SubmissionResponse
	if err := copier.CopyWithOption(&submissionsResponse, submissions, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map submissions", err.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Submissions fetched", submissionsResponse, meta)
}

// GetDetailSubmission for get detail
func (ctrl *SubmissionController) GetDetailSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	submissionID, err := uuid.Parse(c.Params("submissionID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid submission ID", err.Error())
	}

	submission, err := ctrl.submissionService.GetSubmissionDetail(ctx, submissionID, user)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Submission not found", err.Error())
	}

	var submissionResponse dto.SubmissionResponse
	if err := copier.CopyWithOption(&submissionResponse, submission, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map submission", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Submission detail fetched", submissionResponse)
}

// CreateSubmission controller
func (ctrl *SubmissionController) CreateSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.CreateSubmissionRequest)
	user := c.Locals("user").(*utils.Claims)

	assignmentIDParam := c.Params("assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid assignment ID", err.Error())
	}

	// Extract file URLs
	var fileURLs []string
	for _, f := range body.SubmissionFiles {
		fileURLs = append(fileURLs, f.FileURL)
	}

	submission, err := ctrl.submissionService.CreateSubmission(ctx, user, assignmentID, body, fileURLs)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create submission", err.Error())
	}

	var submissionResponse dto.SubmissionResponse
	if err := copier.Copy(&submissionResponse, submission); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map submission data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Submission created successfully", submissionResponse)
}

// UpdateSubmission for PATCH
func (ctrl *SubmissionController) UpdateSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.UpdateSubmissionRequest)
	user := c.Locals("user").(*utils.Claims)

	submissionID, err := uuid.Parse(c.Params("submissionID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid submission ID format", err.Error())
	}

	submission, err := ctrl.submissionService.UpdateSubmission(ctx, user, submissionID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update submission", err.Error())
	}

	var submissionResponse dto.SubmissionResponse
	if err := copier.CopyWithOption(&submissionResponse, submission, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return utils.ErrorResponse(c, 500, "Failed to map submission data", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Submission updated successfully", submissionResponse)
}

// DeleteSubmission for DELETE
func (ctrl *SubmissionController) DeleteSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	submissionID, err := uuid.Parse(c.Params("submissionID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid submission ID format", err.Error())
	}

	if err := ctrl.submissionService.DeleteSubmission(ctx, user, submissionID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to delete submission", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Submission deleted successfully", nil)
}

// GetSubmissionGrade untuk lihat nilai & feedback
func (ctrl *SubmissionController) GetSubmissionGrade(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	submissionID, err := uuid.Parse(c.Params("submissionID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid submission ID format", err.Error())
	}

	submission, err := ctrl.submissionService.GetSubmissionGrade(ctx, user, submissionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.SuccessResponse(c, fiber.StatusNotFound, "Submission grade not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Failed to get submission grade", err.Error())
	}

	var gradeResponse dto.SubmissionGradeResponse
	if err := copier.CopyWithOption(&gradeResponse, submission, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map submission grade data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Submission grade fetched successfully", gradeResponse)
}

// GradeSubmission for post
func (ctrl *SubmissionController) GradeSubmission(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	body := c.Locals("body").(*dto.GradeSubmissionRequest)
	if user.Role != string(models.RoleTypeGuru) {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Only teachers can grade submissions", nil)
	}

	submissionID, err := uuid.Parse(c.Params("submissionID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid submission ID", err.Error())
	}

	grade, err := ctrl.submissionService.GradeSubmission(ctx, user, submissionID, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to grade submission", err.Error())
	}

	var gradeResponse dto.SubmissionGradeResponse
	if err := copier.CopyWithOption(&gradeResponse, grade, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map submission grade data", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Submission graded successfully", gradeResponse)
}

// GenerateGradesExcel controller
func (ctrl *SubmissionController) GenerateGradesExcel(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	assignmentID, err := uuid.Parse(c.Params("assignmentID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid assignment ID", err.Error())
	}

	f, filename, err := ctrl.submissionService.GenerateGradesExcel(ctx, user, assignmentID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate grades excel", err.Error())
	}

	// Simpan sementara di memory
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to write excel buffer", err.Error())
	}

	// Kirim sebagai file download
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.SendStream(buffer)
}

// ImportGradesFromExcel excel
func (ctrl *SubmissionController) ImportGradesFromExcel(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	assignmentID, err := uuid.Parse(c.Params("assignmentID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid assignment ID", err.Error())
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Missing excel file", err.Error())
	}

	if err := ctrl.submissionService.ImportGradesFromExcel(ctx, user, assignmentID, fileHeader); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to import grades from excel", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Grades imported successfully", nil)
}
