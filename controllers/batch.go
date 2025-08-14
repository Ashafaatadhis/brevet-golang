package controllers

import (
	"brevet-api/dto"
	"brevet-api/helpers"
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
	batchService   services.IBatchService
	meetingService services.IMeetingService
	courseService  services.ICourseService
	db             *gorm.DB
}

// NewBatchController creates a new BatchController
func NewBatchController(batchService services.IBatchService, meetingService services.IMeetingService, courseService services.ICourseService, db *gorm.DB) *BatchController {
	return &BatchController{
		batchService:   batchService,
		meetingService: meetingService,
		courseService:  courseService,
		db:             db,
	}
}

// GetAllBatches retrieves a list of batches with pagination and filtering options
func (ctrl *BatchController) GetAllBatches(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("GetAllBatches handler called")
	opts := utils.ParseQueryOptions(c)

	batches, total, err := ctrl.batchService.GetAllFilteredBatches(ctx, opts)
	if err != nil {
		log.WithError(err).Error("Gagal mengambil data batch")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch batches", err.Error())
	}

	batchesResponse := make([]dto.BatchResponse, 0)

	// Loop dan map manual
	for _, batch := range batches {
		var res dto.BatchResponse

		if err := copier.CopyWithOption(&res, batch, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			log.WithError(err).Error("Gagal mapping batch")
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		if err := copier.CopyWithOption(&res.Days, batch.BatchDays, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			log.WithError(err).Error("Gagal mapping batch days")
			return utils.ErrorResponse(c, 500, "Failed to map batch data", err.Error())
		}

		batchesResponse = append(batchesResponse, res)
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	log.WithField("total_batches", total).Info("Berhasil mengambil data batch")
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Batches fetched", batchesResponse, meta)
}

// GetBatchBySlug retrieves a batch by its slug (ID)
func (ctrl *BatchController) GetBatchBySlug(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("GetBatchBySlug handler called")

	slugParam := c.Params("slug")
	log = log.WithField("slug", slugParam)

	batch, err := ctrl.batchService.GetBatchBySlug(ctx, slugParam)
	if err != nil {
		log.WithError(err).Warn("Batch tidak ditemukan")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Batch Doesn't Exist", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping batch")
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	if copyErr := copier.CopyWithOption(&batchResponse.Days, batch.BatchDays, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping batch days")
		return utils.ErrorResponse(c, 500, "Failed to map batch day", copyErr.Error())
	}
	log.Info("Batch berhasil diambil")
	return utils.SuccessResponse(c, fiber.StatusOK, "Batch fetched", batchResponse)
}

// CreateBatch handles the creation of a new batch
func (ctrl *BatchController) CreateBatch(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("CreateBatch handler called")
	body := c.Locals("body").(*dto.CreateBatchRequest)
	user := c.Locals("user").(*utils.Claims)

	log = log.WithField("user_id", user.UserID)

	courseIDParam := c.Params("courseId")
	courseID, err := uuid.Parse(courseIDParam)
	log = log.WithField("course_id", courseID)
	if err != nil {
		log.WithError(err).Warn("UUID courseId tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	batch, err := ctrl.batchService.CreateBatch(ctx, courseID, body)
	if err != nil {
		log.WithError(err).Error("Gagal membuat batch")
		return utils.ErrorResponse(c, 400, "Gagal membuat batch", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping batch")
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}

	if copyErr := copier.CopyWithOption(&batchResponse.Days, batch.BatchDays, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping batch")
		return utils.ErrorResponse(c, 500, "Failed to map batch day", copyErr.Error())
	}
	log.WithField("batch_id", batch.ID).Info("Batch berhasil dibuat")
	return utils.SuccessResponse(c, 201, "Sukses membuat batch", batchResponse)
}

// UpdateBatch updates an existing batch with the provided details
func (ctrl *BatchController) UpdateBatch(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("UpdateBatch handler called")
	user := c.Locals("user").(*utils.Claims)

	log = log.WithField("user_id", user.UserID)
	idParam := c.Params("id")
	log = log.WithField("batch_id", idParam)
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.WithError(err).Warn("UUID batch tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	body := c.Locals("body").(*dto.UpdateBatchRequest)

	batch, err := ctrl.batchService.UpdateBatch(ctx, id, body)
	if err != nil {
		log.WithError(err).Error("Gagal update batch")
		return utils.ErrorResponse(c, 400, "Failed to update batch", err.Error())
	}

	var batchResponse dto.BatchResponse
	if copyErr := copier.CopyWithOption(&batchResponse, batch, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping batch")
		return utils.ErrorResponse(c, 500, "Failed to map batch data", copyErr.Error())
	}
	log.Info("Batch berhasil diupdate")
	return utils.SuccessResponse(c, 200, "Batch updated successfully", batchResponse)
}

// DeleteBatch deletes a batch by its ID
func (ctrl *BatchController) DeleteBatch(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("DeleteBatch handler called")
	user := c.Locals("user").(*utils.Claims)

	log = log.WithField("user_id", user.UserID)

	idParam := c.Params("id")
	log = log.WithField("batch_id", idParam)
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.WithError(err).Warn("UUID batch tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.batchService.DeleteBatch(ctx, id); err != nil {
		log.WithError(err).Error("Gagal menghapus batch")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete batch", err.Error())
	}
	log.Info("Batch berhasil dihapus")
	return utils.SuccessResponse(c, fiber.StatusOK, "Batch deleted successfully", nil)
}

// GetBatchByCourseSlug this function for get batch by course slug
func (ctrl *BatchController) GetBatchByCourseSlug(c *fiber.Ctx) error {
	ctx := c.UserContext()

	courseSlug := c.Params("courseSlug")

	opts := utils.ParseQueryOptions(c)

	course, err := ctrl.courseService.GetCourseBySlug(ctx, courseSlug)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Course not found", err.Error())
	}

	batches, total, err := ctrl.batchService.GetBatchByCourseSlug(ctx, course.ID, opts)
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
	ctx := c.UserContext()

	user := c.Locals("user").(*utils.Claims)
	opts := utils.ParseQueryOptions(c)

	var batches []models.Batch
	var total int64
	var err error

	switch user.Role {
	case string(models.RoleTypeSiswa):
		batches, total, err = ctrl.batchService.GetBatchesPurchasedByUser(ctx, user.UserID, opts)
	case string(models.RoleTypeGuru):
		batches, total, err = ctrl.batchService.GetBatchesTaughtByGuru(ctx, user.UserID, opts)
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

// GetMyMeetings this function for mymeetings controller
func (ctrl *BatchController) GetMyMeetings(c *fiber.Ctx) error {
	ctx := c.UserContext()

	user := c.Locals("user").(*utils.Claims)
	batchSlug := c.Params("batchSlug")
	opts := utils.ParseQueryOptions(c)

	var meetings []models.Meeting
	var total int64
	var err error

	switch user.Role {
	case string(models.RoleTypeSiswa):
		meetings, total, err = ctrl.meetingService.GetMeetingsPurchasedByUser(ctx, user.UserID, batchSlug, opts)
	case string(models.RoleTypeGuru):
		meetings, total, err = ctrl.meetingService.GetMeetingsTaughtByTeacher(ctx, user.UserID, batchSlug, opts)
	default:
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", "Hanya siswa dan guru yang dapat melihat meetings ini")
	}

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data batch", err.Error())
	}

	var meetingsResponse []dto.MeetingResponse

	if err := copier.CopyWithOption(&meetingsResponse, meetings, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", err.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Batch berhasil diambil", meetingsResponse, meta)
}

// GetAllStudents get all students
func (ctrl *BatchController) GetAllStudents(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchSlug := c.Params("batchSlug")

	user := c.Locals("user").(*utils.Claims)
	opts := utils.ParseQueryOptions(c)

	var total int64
	var err error

	students, total, err := ctrl.meetingService.GetStudentsByBatchSlugFiltered(ctx, user, batchSlug, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get students", err.Error())
	}

	var userResponses []dto.UserResponse
	if copyErr := copier.Copy(&userResponses, students); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Meetings fetched", userResponses, meta)

}

// GetProgress for get progress
func (ctrl *BatchController) GetProgress(c *fiber.Ctx) error {
	ctx := c.UserContext()

	user := c.Locals("user").(*utils.Claims)
	batchID, err := uuid.Parse(c.Params("batchID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid batch ID", err.Error())
	}

	progress, err := ctrl.batchService.CalculateProgress(ctx, batchID, user.UserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to get progress", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Progress fetched successfully", fiber.Map{
		"progress_percent": progress,
	})
}
