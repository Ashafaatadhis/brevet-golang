package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

// UserController represents the authentication controller
type UserController struct {
	userService *services.UserService
}

// NewUserController is a constructor for UserController
func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

// GetAllUsers retrieves all users
func (ctrl *UserController) GetAllUsers(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	users, total, err := ctrl.userService.GetAllFilteredUsers(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	var usersResponse []dto.UserResponse
	if copyErr := copier.Copy(&usersResponse, users); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map user data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Users fetched", usersResponse, meta)
}

// GetProfile retrieves the profile of the authenticated user
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	token := c.Locals("user").(*utils.Claims)

	user, err := ctrl.userService.GetUserByID(token.UserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	var userResponse dto.UserResponse
	if copyErr := copier.Copy(&userResponse, user); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map user data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile fetched", userResponse)
}
