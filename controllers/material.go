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

// MaterialController handles material controller
type MaterialController struct {
	materialService services.IMaterialService
	db              *gorm.DB
}

// NewMaterialController creates a new instance of MaterialController
func NewMaterialController(materialService services.IMaterialService, db *gorm.DB) *MaterialController {
	return &MaterialController{
		materialService: materialService,
		db:              db,
	}
}

// GetAllMaterials retrieves a list of materials with pagination and filtering options
func (ctrl *MaterialController) GetAllMaterials(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)

	materials, total, err := ctrl.materialService.GetAllFilteredMaterial(ctx, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch materials", err.Error())
	}

	var materialsResponse []dto.MaterialResponse
	if copyErr := copier.Copy(&materialsResponse, materials); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map material data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Materials fetched", materialsResponse, meta)
}

// GetAllMaterialByMeetingID retrieves a list of materials with pagination and filtering options
func (ctrl *MaterialController) GetAllMaterialByMeetingID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	opts := utils.ParseQueryOptions(c)
	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)

	user := c.Locals("user").(*utils.Claims)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}
	materials, total, err := ctrl.materialService.GetAllFilteredMaterialsByMeetingID(ctx, meetingID, user, opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch material", err.Error())
	}

	var materialsResponse []dto.MaterialResponse
	if copyErr := copier.Copy(&materialsResponse, materials); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map material data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Materials fetched", materialsResponse, meta)
}

// GetMaterialByID retrieves a single material by its ID
func (ctrl *MaterialController) GetMaterialByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	materialIDParam := c.Params("materialID")
	materialID, err := uuid.Parse(materialIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	user := c.Locals("user").(*utils.Claims)

	material, err := ctrl.materialService.GetMaterialByID(ctx, user, materialID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Material not found", err.Error())
	}

	var materialResponse dto.MaterialResponse
	if copyErr := copier.Copy(&materialResponse, material); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map material data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Material retrieved successfully", materialResponse)
}

// CreateMaterial creates a new material with the provided details
func (ctrl *MaterialController) CreateMaterial(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.CreateMaterialRequest)
	user := c.Locals("user").(*utils.Claims)

	meetingIDParam := c.Params("meetingID")
	meetingID, err := uuid.Parse(meetingIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	material, err := ctrl.materialService.CreateMaterial(ctx, user, meetingID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to create material", err.Error())
	}

	var materialResponse dto.MaterialResponse
	if copyErr := copier.Copy(&materialResponse, material); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map material data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Material created successfully", materialResponse)
}

// UpdateMaterial updates an existing material
func (ctrl *MaterialController) UpdateMaterial(c *fiber.Ctx) error {
	ctx := c.UserContext()
	body := c.Locals("body").(*dto.UpdateMaterialRequest)
	user := c.Locals("user").(*utils.Claims)

	materialIDParam := c.Params("materialID")
	materialID, err := uuid.Parse(materialIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	material, err := ctrl.materialService.UpdateMaterial(ctx, user, materialID, body)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to update material", err.Error())
	}

	var materialResponse dto.MaterialResponse
	if copyErr := copier.Copy(&materialResponse, material); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map material data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Material updated successfully", materialResponse)
}

// DeleteMaterial deletes an existing material and its related files
func (ctrl *MaterialController) DeleteMaterial(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user := c.Locals("user").(*utils.Claims)

	materialIDParam := c.Params("materialID")
	materialID, err := uuid.Parse(materialIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	if err := ctrl.materialService.DeleteMaterial(ctx, user, materialID); err != nil {
		return utils.ErrorResponse(c, 400, "Failed to delete material", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Material deleted successfully", nil)
}
