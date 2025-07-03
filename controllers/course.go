package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CourseController handles course-related operations
type CourseController struct {
	courseService *services.CourseService
	db            *gorm.DB
}

// NewCourseController creates a new CourseController
func NewCourseController(courseService *services.CourseService, db *gorm.DB) *CourseController {
	return &CourseController{
		courseService: courseService,
		db:            db,
	}
}

// GetAllCourses retrieves a list of courses with pagination and filtering options
func (ctrl *CourseController) GetAllCourses(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	courses, total, err := ctrl.courseService.GetAllFilteredCourses(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch courses", err.Error())
	}

	var coursesResponse []dto.CourseResponse
	if copyErr := copier.Copy(&coursesResponse, courses); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Courses fetched", coursesResponse, meta)
}

// GetCourseBySlug retrieves a course by its slug (ID)
func (ctrl *CourseController) GetCourseBySlug(c *fiber.Ctx) error {
	slugParam := c.Params("slug")

	course, err := ctrl.courseService.GetCourseBySlug(slugParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Course Doesn't Exist", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Course fetched", courseResponse)
}

// CreateCourse creates a new course with the provided details
func (ctrl *CourseController) CreateCourse(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateCourseRequest)

	course, err := ctrl.courseService.CreateCourse(body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to create course", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Course created successfully", courseResponse)
}

// UpdateCourse updates an existing course with the provided details
func (ctrl *CourseController) UpdateCourse(c *fiber.Ctx) error {
	log.Print("UpdateCourse called\n")
	fmt.Print("TAIKKKK")
	idParam := c.Params("id")
	fmt.Print(idParam, "ID Param\n")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	body := c.Locals("body").(*dto.UpdateCourseRequest)

	course, err := ctrl.courseService.UpdateCourse(id, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update course", err.Error())
	}

	var courseResponse dto.CourseResponse
	if copyErr := copier.Copy(&courseResponse, course); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map course data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Course updated successfully", courseResponse)
}

// DeleteCourse deletes a course by its ID
func (ctrl *CourseController) DeleteCourse(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.courseService.DeleteCourse(id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete course", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Course deleted successfully", nil)
}
