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

// BlogController handles blog-related operations
type BlogController struct {
	blogService services.IBlogService
	db          *gorm.DB
}

// NewBlogController creates a new BlogController
func NewBlogController(blogService services.IBlogService, db *gorm.DB) *BlogController {
	return &BlogController{
		blogService: blogService,
		db:          db,
	}
}

// GetAllBlogs retrieves a list of blogs with pagination and filtering options
func (ctrl *BlogController) GetAllBlogs(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)

	blogs, total, err := ctrl.blogService.GetAllFilteredBlogs(ctx, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch blogs", err.Error())
	}

	var blogsResponse []dto.BlogResponse
	if copyErr := copier.Copy(&blogsResponse, blogs); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map blog data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Blogs fetched", blogsResponse, meta)
}

// GetBlogBySlug retrieves a blog by its slug (ID)
func (ctrl *BlogController) GetBlogBySlug(c *fiber.Ctx) error {
	ctx := c.UserContext()
	slugParam := c.Params("slug")

	blog, err := ctrl.blogService.GetBlogBySlug(ctx, slugParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Blog Doesn't Exist", err.Error())
	}

	var blogResponse dto.BlogResponse
	if copyErr := copier.Copy(&blogResponse, blog); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map blog data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog fetched", blogResponse)
}

// CreateBlog handles the creation of a new blog
func (ctrl *BlogController) CreateBlog(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.CreateBlogRequest)

	blog, err := ctrl.blogService.CreateBlog(ctx, body)
	if err != nil {

		return utils.ErrorResponse(c, 400, "Gagal membuat blog", err.Error())
	}
	var blogResponse dto.BlogResponse
	if copyErr := copier.Copy(&blogResponse, blog); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map blog data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Sukses membuat blog", blogResponse)
}

// UpdateBlog updates an existing blog with the provided details
func (ctrl *BlogController) UpdateBlog(c *fiber.Ctx) error {
	ctx := c.UserContext()
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	body := c.Locals("body").(*dto.UpdateBlogRequest)

	blog, err := ctrl.blogService.UpdateBlog(ctx, id, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update blog", err.Error())
	}

	var blogResponse dto.BlogResponse
	if copyErr := copier.Copy(&blogResponse, blog); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map blog data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Blog updated successfully", blogResponse)
}

// DeleteBlog deletes a blog by its ID
func (ctrl *BlogController) DeleteBlog(c *fiber.Ctx) error {
	ctx := c.UserContext()
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.blogService.DeleteBlog(ctx, id); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete blog", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Blog deleted successfully", nil)
}
