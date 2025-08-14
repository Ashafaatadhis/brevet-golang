package controllers

import (
	"brevet-api/dto"
	"brevet-api/helpers"
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
	userService services.IUserService
	authService services.IAuthService
	db          *gorm.DB
}

// NewUserController is a constructor for UserController
func NewUserController(userService services.IUserService, authService services.IAuthService, db *gorm.DB) *UserController {
	return &UserController{userService: userService, authService: authService, db: db}
}

// GetAllUsers retrieves all users
func (ctrl *UserController) GetAllUsers(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("GetAllUsers handler called")
	opts := utils.ParseQueryOptions(c)

	users, total, err := ctrl.userService.GetAllFilteredUsers(ctx, opts)
	if err != nil {
		log.WithError(err).Error("Gagal mengambil daftar user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", err.Error())
	}

	var usersResponse []dto.UserResponse
	if copyErr := copier.Copy(&usersResponse, users); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data user")
		return utils.ErrorResponse(c, 500, "Failed to map user data", copyErr.Error())
	}
	log.WithField("total", total).Info("User list fetched successfully")
	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Users fetched", usersResponse, meta)
}

// GetUserByID is represent to get user by id
func (ctrl *UserController) GetUserByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	idParam := c.Params("id")
	log = log.WithField("user_id", idParam)
	log.Info("GetUserByID handler called")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.WithError(err).Warn("UUID tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	user, err := ctrl.userService.GetUserByID(ctx, id) // pastikan parameternya uuid.UUID

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User tidak ditemukan")
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		log.WithError(err).Error("Gagal mengambil user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch user", err.Error())
	}

	var userResponse dto.UserResponse
	if copyErr := copier.Copy(&userResponse, user); copyErr != nil {
		log.WithError(copyErr).Error("Gagal mapping data user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map user data", copyErr.Error())
	}
	log.Info("User fetched successfully")
	return utils.SuccessResponse(c, fiber.StatusOK, "User fetched", userResponse)
}

// GetProfile retrieves the profile of the authenticated user
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	claims := c.Locals("user").(*utils.Claims)
	log = log.WithField("user_id", claims.UserID)
	log.Info("GetProfile handler called")
	userResp, err := ctrl.userService.GetProfileResponseByID(ctx, claims.UserID)
	if err != nil {
		log.WithError(err).Warn("User tidak ditemukan saat ambil profil")
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}
	log.Info("Profil user berhasil diambil")
	return utils.SuccessResponse(c, fiber.StatusOK, "Profile fetched", userResp)
}

// CreateUserWithProfile is for create user
func (ctrl *UserController) CreateUserWithProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("CreateUserWithProfile handler called")
	body := c.Locals("body").(*dto.CreateUserWithProfileRequest)

	// Validasi minimum (cukup di sini)
	if body.RoleType == models.RoleTypeSiswa {
		if body.NIK == nil {
			log.Warn("NIK wajib diisi untuk RoleTypeSiswa")
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Mahasiswa wajib mengisi NIK", nil)
		}
		if (body.NIM == nil && body.NIMProof != nil) || (body.NIM != nil && body.NIMProof == nil) {
			log.Warn("NIM dan bukti harus diisi bersamaan")
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "NIM dan bukti NIM harus diisi bersamaan", nil)
		}
	}

	// Delegasikan ke service
	userResp, err := ctrl.userService.CreateUserWithProfile(ctx, body)
	if err != nil {
		log.WithError(err).Error("Gagal membuat user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat user", err.Error())
	}
	log.WithField("user_id", userResp.ID).Info("User berhasil dibuat")
	return utils.SuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", userResp)
}

// UpdateUserWithProfile untuk controler update user
func (ctrl *UserController) UpdateUserWithProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	userIDStr := c.Params("id")
	log = log.WithField("user_id", userIDStr)
	log.Info("UpdateUserWithProfile handler called")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.WithError(err).Warn("UUID tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	body := c.Locals("body").(*dto.UpdateUserWithProfileRequest)

	// Delegasikan semua ke service
	userResp, err := ctrl.userService.UpdateUserWithProfile(ctx, userID, body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User tidak ditemukan saat update")
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		log.WithError(err).Error("Gagal update user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui user", err.Error())
	}
	log.Info("User berhasil diperbarui")
	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil diperbarui", userResp)
}

// DeleteUserByID for delete user controller
func (ctrl *UserController) DeleteUserByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	userIDStr := c.Params("id")
	log = log.WithField("user_id", userIDStr)
	log.Info("DeleteUserByID handler called")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.WithError(err).Warn("UUID tidak valid")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	// Delegasi ke service
	if err := ctrl.userService.DeleteUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User tidak ditemukan saat hapus")
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		log.WithError(err).Error("Gagal menghapus user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus user", err.Error())
	}
	log.Info("User berhasil dihapus")
	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
}

// UpdateMyProfile updates the profile of the authenticated user
func (ctrl *UserController) UpdateMyProfile(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)

	claims := c.Locals("user").(*utils.Claims)
	userID := claims.UserID
	log = log.WithField("user_id", claims.UserID)
	log.Info("UpdateMyProfile handler called")

	body := c.Locals("body").(*dto.UpdateMyProfile)

	// Delegasikan semuanya ke service
	userResp, err := ctrl.userService.UpdateMyProfile(ctx, userID, body)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("User tidak ditemukan saat update profil")
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		log.WithError(err).Error("Gagal update profil user")
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal update profil", err.Error())
	}
	log.Info("Profil user berhasil diperbarui")
	return utils.SuccessResponse(c, fiber.StatusOK, "Profil berhasil diperbarui", userResp)
}
