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

// GroupController is a controller for group-related operations
type GroupController struct {
	groupService *services.GroupService
	db           *gorm.DB
}

// NewGroupController creates a new GroupController instance
func NewGroupController(groupService *services.GroupService, db *gorm.DB) *GroupController {
	return &GroupController{groupService: groupService, db: db}
}

// GetAllGroups handles the request to get all groups
func (ctrl *GroupController) GetAllGroups(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	groups, total, err := ctrl.groupService.GetAllGroups(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	var groupsResponse []dto.GroupResponse
	if copyErr := copier.Copy(&groupsResponse, groups); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map group data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Groups fetched", groupsResponse, meta)
}

// GetGroupByID handles the request to get a group by its ID
func (ctrl *GroupController) GetGroupByID(c *fiber.Ctx) error {
	groupIDStr := c.Params("id")
	if groupIDStr == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Group ID is required", "Group ID cannot be empty")
	}

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid Group ID format", "Group ID must be a valid UUID")
	}

	group, err := ctrl.groupService.GetGroupByID(groupID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Group not found", "No group found with the given ID")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch group", err.Error())
	}

	var groupResponse dto.GroupResponse
	if copyErr := copier.Copy(&groupResponse, group); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map group data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Group fetched successfully", groupResponse)
}
