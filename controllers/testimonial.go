package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

// TestimonialController controller
type TestimonialController struct {
	testimonialService services.ITestimonialService
}

// NewTestimonialController init
func NewTestimonialController(testimonialService services.ITestimonialService) *TestimonialController {
	return &TestimonialController{testimonialService: testimonialService}
}

// GetAllFiltered get all filtered
func (ctrl *TestimonialController) GetAllFiltered(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)

	testimonials, total, err := ctrl.testimonialService.GetAllFiltered(ctx, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch testimonials", err.Error())
	}

	var resp []dto.TestimonialResponse
	if copyErr := copier.Copy(&resp, testimonials); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map testimonials", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Testimonials fetched", resp, meta)
}

// GetByBatchIDFiltered get all filtered
func (ctrl *TestimonialController) GetByBatchIDFiltered(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)

	batchSlug := c.Params("batchSlug")

	testimonials, total, err := ctrl.testimonialService.GetByBatchSlugFiltered(ctx, batchSlug, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch testimonials", err.Error())
	}

	var resp []dto.TestimonialResponse
	if copyErr := copier.Copy(&resp, testimonials); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map testimonials", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Testimonials fetched", resp, meta)
}

// GetByID get detail
func (ctrl *TestimonialController) GetByID(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Ambil ID dari params
	idStr := c.Params("testimonialID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid testimonial id", err.Error())
	}

	testimonial, err := ctrl.testimonialService.GetByID(ctx, id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "testimonial not found", err.Error())
	}

	var resp dto.TestimonialResponse
	copier.Copy(&resp, testimonial)
	return utils.SuccessResponse(c, fiber.StatusOK, "Testimonial fetched", resp)
}

// Create TestimonialController create
func (ctrl *TestimonialController) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	req := c.Locals("body").(*dto.CreateTestimonialRequest)

	batchID, err := uuid.Parse(c.Params("batchID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid batch ID", err.Error())
	}

	testimonial, err := ctrl.testimonialService.Create(ctx, req, batchID, user.UserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create testimonial", err.Error())
	}

	var resp dto.TestimonialResponse
	copier.Copy(&resp, testimonial)
	return utils.SuccessResponse(c, fiber.StatusCreated, "Testimonial created", resp)
}

// Update testimonial
func (ctrl *TestimonialController) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)
	req := c.Locals("body").(*dto.UpdateTestimonialRequest)
	id, err := uuid.Parse(c.Params("testimonialID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid testimonial ID", err.Error())
	}

	testimonial, err := ctrl.testimonialService.Update(ctx, id, req, user.UserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to update testimonial", err.Error())
	}

	var resp dto.TestimonialResponse
	copier.Copy(&resp, testimonial)
	return utils.SuccessResponse(c, fiber.StatusOK, "Testimonial updated", resp)
}

// Delete testimonial
func (ctrl *TestimonialController) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	id, err := uuid.Parse(c.Params("testimonialID"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid testimonial ID", err.Error())
	}

	if err := ctrl.testimonialService.Delete(ctx, id, user.UserID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete testimonial", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Testimonial deleted", nil)
}
