package controllers

import (
	"brevet-api/dto"
	"brevet-api/models"
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

// GetMyBatches this function for mybatches controller
func (ctrl *BatchController) GetMyBatches(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.Claims)
	opts := utils.ParseQueryOptions(c)

	var batches []models.Batch
	var total int64
	var err error

	switch user.Role {
	case string(models.RoleTypeSiswa):
		batches, total, err = ctrl.batchService.GetBatchesPurchasedByUser(user.UserID, opts)
	case string(models.RoleTypeGuru):
		batches, total, err = ctrl.batchService.GetBatchesTaughtByGuru(user.UserID, opts)
	default:
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", "Hanya siswa dan guru yang dapat melihat batch ini")
	}

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data batch", err.Error())
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

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Batch berhasil diambil", batchesResponse, meta)
}
