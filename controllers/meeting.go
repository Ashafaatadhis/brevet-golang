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

// MeetingController handles meeting-related operations
type MeetingController struct {
	meetingService *services.MeetingService
	db             *gorm.DB
}

// NewMeetingController creates a new NewMeetingController
func NewMeetingController(meetingService *services.MeetingService, db *gorm.DB) *MeetingController {
	return &MeetingController{
		meetingService: meetingService,
		db:             db,
	}
}

// GetAllMeetings retrieves a list of meetings with pagination and filtering options
func (ctrl *MeetingController) GetAllMeetings(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	meetings, total, err := ctrl.meetingService.GetAllFilteredMeetings(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch meetings", err.Error())
	}

	var meetingResponse []dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meetings); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Meetings fetched", meetingResponse, meta)
}

// GetMeetingsByBatchSlug retrieves a list of meetings for a specific batch
func (ctrl *MeetingController) GetMeetingsByBatchSlug(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	batchSlug := c.Params("batchSlug")
	meetings, total, err := ctrl.meetingService.GetMeetingsByBatchSlug(batchSlug, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch meetings", err.Error())
	}

	var meetingResponse []dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meetings); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Meetings fetched", meetingResponse, meta)
}

// GetMeetingByID is controller that retrieves meeting by them id
func (ctrl *MeetingController) GetMeetingByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	meeting, err := ctrl.meetingService.GetMeetingByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Meeting Doesn't Exist", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Meeting fetched", meetingResponse)
}

// CreateMeeting is for create meeting
func (ctrl *MeetingController) CreateMeeting(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateMeetingRequest)
	idParam := c.Params("batchID")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	meeting, err := ctrl.meetingService.CreateMeeting(id, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create meeting", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Meeting created successfully", meetingResponse)
}

// UpdateMeeting is for update meeting
func (ctrl *MeetingController) UpdateMeeting(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.UpdateMeetingRequest)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	meeting, err := ctrl.meetingService.UpdateMeeting(id, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update meeting", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Meeting updated successfully", meetingResponse)
}

// DeleteMeeting deletes a meeting by its ID
func (ctrl *MeetingController) DeleteMeeting(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.meetingService.DeleteMeeting(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete meeting", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Meeting deleted successfully", nil)
}

// AddTeachersToMeeting is function to add teacher to meeting
func (ctrl *MeetingController) AddTeachersToMeeting(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.AssignTeachersRequest)
	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	meeting, err := ctrl.meetingService.AddTeachersToMeeting(meetingID, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan guru ke meeting", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Guru berhasil ditambahkan", meetingResponse)
}

// UpdateTeachersToMeeting is function to update teacher to meeting
func (ctrl *MeetingController) UpdateTeachersToMeeting(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.AssignTeachersRequest)
	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	meeting, err := ctrl.meetingService.UpdateTeachersInMeeting(meetingID, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate guru ke meeting", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Guru berhasil diupdate", meetingResponse)
}

// DeleteTeachersToMeeting is function to delete teacher to meeting
func (ctrl *MeetingController) DeleteTeachersToMeeting(c *fiber.Ctx) error {

	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	teacherIDParam := c.Params("teacherID")
	teacherID, err := uuid.Parse(teacherIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid teacher ID", err.Error())
	}

	meeting, err := ctrl.meetingService.RemoveTeachersFromMeeting(meetingID, teacherID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate guru ke meeting", err.Error())
	}

	var meetingResponse dto.MeetingResponse
	if copyErr := copier.Copy(&meetingResponse, meeting); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Guru berhasil diupdate", meetingResponse)
}

// GetTeachersByMeetingIDFiltered controller that get teachers by meta such a pagination
func (ctrl *MeetingController) GetTeachersByMeetingIDFiltered(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)
	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid meeting ID", err.Error())
	}

	teachers, total, err := ctrl.meetingService.GetTeachersByMeetingIDFiltered(meetingID, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch teachers", err.Error())
	}

	var userResponses []dto.UserResponse
	if copyErr := copier.Copy(&userResponses, teachers); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map meeting data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Meetings fetched", userResponses, meta)
}
