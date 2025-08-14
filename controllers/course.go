package controllers

import (
	"brevet-api/dto"
	"brevet-api/helpers"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CourseController handles course-related operations
type CourseController struct {
	courseService services.ICourseService
	db            *gorm.DB
}

// NewCourseController creates a new CourseController
func NewCourseController(courseService services.ICourseService, db *gorm.DB) *CourseController {
	return &CourseController{
		courseService: courseService,
		db:            db,
	}
}

// GetAllCourses retrieves a list of courses with pagination and filtering options
func (ctrl *CourseController) GetAllCourses(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("GetAllCourses handler called")
	opts := utils.ParseQueryOptions(c)

	courses, total, err := ctrl.courseService.GetAllFilteredCourses(ctx, opts)
	if err != nil {
		log.WithError(err).Error("Gagal mengambil daftar kursus")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch courses", err.Error())
	}

	var coursesResponse []dto.CourseResponse
	if copyErr := copier.Copy(&coursesResponse, courses); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data kursus")
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}
	log.WithField("total", total).Info("Daftar kursus berhasil diambil")
	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Courses fetched", coursesResponse, meta)
}

// GetCourseBySlug retrieves a course by its slug (ID)
func (ctrl *CourseController) GetCourseBySlug(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("GetCourseBySlug handler called")
	slugParam := c.Params("slug")
	log = log.WithField("slug", slugParam)
	course, err := ctrl.courseService.GetCourseBySlug(ctx, slugParam)
	if err != nil {
		log.WithError(err).Warn("Kursus tidak ditemukan")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Course Doesn't Exist", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data kursus")
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}
	log.Info("Kursus berhasil ditemukan")
	return utils.SuccessResponse(c, fiber.StatusOK, "Course fetched", courseResponse)
}

// CreateCourse creates a new course with the provided details
func (ctrl *CourseController) CreateCourse(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	body := c.Locals("body").(*dto.CreateCourseRequest)
	user := c.Locals("user").(*utils.Claims)
	log = log.WithField("user_id", user.ID)
	log.Info("CreateCourse handler called")
	course, err := ctrl.courseService.CreateCourse(ctx, body)
	if err != nil {
		log.WithError(err).Error("Gagal membuat kursus")
		return utils.ErrorResponse(c, 400, "Failed to create course", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data kursus")
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}
	log.WithField("course_id", course.ID).Info("Kursus berhasil dibuat")
	return utils.SuccessResponse(c, 201, "Course created successfully", courseResponse)
}

// UpdateCourse updates an existing course with the provided details
func (ctrl *CourseController) UpdateCourse(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	user := c.Locals("user").(*utils.Claims)
	log = log.WithField("user_id", user.ID)
	idParam := c.Params("id")

	log = log.WithField("course_id", idParam)

	log.Info("UpdateCourse handler called")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.WithError(err).Warn("UUID tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	body := c.Locals("body").(*dto.UpdateCourseRequest)

	course, err := ctrl.courseService.UpdateCourse(ctx, id, body)
	if err != nil {
		log.WithError(err).Error("Gagal memperbarui kursus")
		return utils.ErrorResponse(c, 400, "Failed to update course", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data kursus")
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}
	log.Info("Kursus berhasil diperbarui")
	return utils.SuccessResponse(c, 200, "Course updated successfully", courseResponse)
}

// DeleteCourse deletes a course by its ID
func (ctrl *CourseController) DeleteCourse(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	user := c.Locals("user").(*utils.Claims)
	log = log.WithField("user_id", user.ID)

	idParam := c.Params("id")
	log = log.WithField("course_id", idParam)
	log.Info("DeleteCourse handler called")

	id, err := uuid.Parse(idParam)
	if err != nil {
		log.WithError(err).Warn("UUID tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.courseService.DeleteCourse(ctx, id); err != nil {
		log.WithError(err).Error("Gagal menghapus kursus")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete course", err.Error())
	}
	log.Info("Kursus berhasil dihapus")
	return utils.SuccessResponse(c, fiber.StatusOK, "Course deleted successfully", nil)
}
