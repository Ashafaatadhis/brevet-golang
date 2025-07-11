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

// BatchController handles batch-related operations
type BatchController struct {
	batchService  *services.BatchService
	courseService *services.CourseService
	db            *gorm.DB
}

// NewBatchController creates a new BatchController
func NewBatchController(batchService *services.BatchService, courseService *services.CourseService, db *gorm.DB) *BatchController {
	return &BatchController{
		batchService:  batchService,
		courseService: courseService,
		db:            db,
	}
}

// GetAllBatches retrieves a list of batches with pagination and filtering options
func (ctrl *BatchController) GetAllBatches(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	batches, total, err := ctrl.batchService.GetAllFilteredBatches(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch batches", err.Error())
	}

	var batchesResponse []dto.BatchResponse

	// Loop dan map manual
	for _, batch := range batches {
		var res dto.BatchResponse

		if err := copier.CopyWithOption(&res, batch, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		if err := copier.CopyWithOption(&res.Days, batch.BatchDays, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		batchesResponse = append(batchesResponse, res)
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Batches fetched", batchesResponse, meta)
}

// GetBatchBySlug retrieves a batch by its slug (ID)
func (ctrl *BatchController) GetBatchBySlug(c *fiber.Ctx) error {
	slugParam := c.Params("slug")

	batch, err := ctrl.batchService.GetBatchBySlug(slugParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Batch Doesn't Exist", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	if copyErr := copier.CopyWithOption(&batchResponse.Days, batch.BatchDays, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch day", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Batch fetched", batchResponse)
}

// CreateBatch handles the creation of a new batch
func (ctrl *BatchController) CreateBatch(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateBatchRequest)

	courseIDParam := c.Params("courseId")
	courseID, err := uuid.Parse(courseIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	batch, err := ctrl.batchService.CreateBatch(courseID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Gagal membuat batch", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	if copyErr := copier.CopyWithOption(&batchResponse.Days, batch.BatchDays, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch day", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Sukses membuat batch", batchResponse)
}

// UpdateBatch updates an existing batch with the provided details
func (ctrl *BatchController) UpdateBatch(c *fiber.Ctx) error {

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	body := c.Locals("body").(*dto.UpdateBatchRequest)

	batch, err := ctrl.batchService.UpdateBatch(id, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update batch", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Batch updated successfully", batchResponse)
}

// DeleteBatch deletes a batch by its ID
func (ctrl *BatchController) DeleteBatch(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.batchService.DeleteBatch(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete batch", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Batch deleted successfully", nil)
}

// ASSIGN TEACHER TO BATCH

// AddTeacherToBatch adds a teacher to a batch
func (ctrl *BatchController) AddTeacherToBatch(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateBatchTeacherRequest)

	batchIDParam := c.Params("batchID")
	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	user, err := ctrl.batchService.AddTeacherToBatch(batchID, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add teacher to batch", err.Error())
	}

	var userResponse dto.UserResponse
	if copyErr := copier.Copy(&userResponse, user); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Sukses membuat batch", userResponse)
}

// GetTeachersByBatch is represent get teacher by batch
func (ctrl *BatchController) GetTeachersByBatch(c *fiber.Ctx) error {
	batchIDParam := c.Params("batchID")
	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	opts := utils.ParseQueryOptions(c)

	teachers, total, err := ctrl.batchService.GetTeacherInBatch(batchID, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch teachers", err.Error())
	}

	var teachersResponse []dto.UserResponse
	if copyErr := copier.Copy(&teachersResponse, teachers); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map teacher data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Teachers fetched", teachersResponse, meta)
}

// RemoveTeacherFromBatch is controller for remove teacher from batch
func (ctrl *BatchController) RemoveTeacherFromBatch(c *fiber.Ctx) error {
	batchIDParam := c.Params("batchID")
	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	userIDParam := c.Params("userID")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.batchService.DeleteTeacherFromBatch(batchID, userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete teacher from batch", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Teacher in batch deleted successfully", nil)

}

// GetBatchByCourseSlug this function for get batch by course slug
func (ctrl *BatchController) GetBatchByCourseSlug(c *fiber.Ctx) error {
	courseSlug := c.Params("courseSlug")

	opts := utils.ParseQueryOptions(c)

	course, err := ctrl.courseService.GetCourseBySlug(courseSlug)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Course not found", err.Error())
	}

	batches, total, err := ctrl.batchService.GetBatchByCourseSlug(course.ID, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch teachers", err.Error())
	}

	var batchesResponse []dto.BatchResponse

	// Loop dan map manual
	for _, batch := range batches {
		var res dto.BatchResponse

		if err := copier.CopyWithOption(&res, batch, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		if err := copier.CopyWithOption(&res.Days, batch.BatchDays, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		batchesResponse = append(batchesResponse, res)
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Batches fetched", batchesResponse, meta)
}
