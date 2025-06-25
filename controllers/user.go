package controllers

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// UserController represents the authentication controller
type UserController struct {
	userService *services.UserService
	authService *services.AuthService
	db          *gorm.DB
}

// NewUserController is a constructor for UserController
func NewUserController(userService *services.UserService, authService *services.AuthService, db *gorm.DB) *UserController {
	return &UserController{userService: userService, authService: authService, db: db}
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

// GetUserByID is represent to get user by id
func (ctrl *UserController) GetUserByID(c *fiber.Ctx) error {
	idParam := c.Params("id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	user, err := ctrl.userService.GetUserByID(id) // pastikan parameternya uuid.UUID

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch user", err.Error())
	}

	var userResponse dto.UserResponse
	if copyErr := copier.Copy(&userResponse, user); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map user data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User fetched", userResponse)
}

// GetProfile retrieves the profile of the authenticated user
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)

	userResp, err := ctrl.userService.GetProfileResponseByID(claims.UserID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profile fetched", userResp)
}

// CreateUserWithProfile is for create user
func (ctrl *UserController) CreateUserWithProfile(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateUserWithProfileRequest)

	// Validasi minimum (cukup di sini)
	if body.RoleType == models.RoleTypeSiswa {
		if body.NIK == nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Mahasiswa wajib mengisi NIK", nil)
		}
		if (body.NIM == nil && body.NIMProof != nil) || (body.NIM != nil && body.NIMProof == nil) {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "NIM dan bukti NIM harus diisi bersamaan", nil)
		}
	}

	// Delegasikan ke service
	userResp, err := ctrl.userService.CreateUserWithProfile(body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", userResp)
}

// UpdateUserWithProfile untuk controler update user
func (ctrl *UserController) UpdateUserWithProfile(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	body := c.Locals("body").(*dto.UpdateUserWithProfileRequest)

	// Delegasikan semua ke service
	userResp, err := ctrl.userService.UpdateUserWithProfile(userID, body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil diperbarui", userResp)
}

// DeleteUserByID for delete user controller
func (ctrl *UserController) DeleteUserByID(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	// Delegasi ke service
	if err := ctrl.userService.DeleteUserByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
}

// UpdateMyProfile updates the profile of the authenticated user
func (ctrl *UserController) UpdateMyProfile(c *fiber.Ctx) error {
	claims := c.Locals("user").(*utils.Claims)
	userID := claims.UserID

	body := c.Locals("body").(*dto.UpdateMyProfile)

	// Delegasikan semuanya ke service
	userResp, err := ctrl.userService.UpdateMyProfile(userID, body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal update profil", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Profil berhasil diperbarui", userResp)
}
