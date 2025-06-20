package controllers

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"errors"
	"fmt"

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

// CreateUserWithProfile is for create user
func (ctrl *UserController) CreateUserWithProfile(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreateUserWithProfileRequest)

	tx := ctrl.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Cek apakah email/phone sudah digunakan
	if !ctrl.authService.IsEmailUnique(tx, body.Email) || !ctrl.authService.IsPhoneUnique(tx, body.Phone) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Email atau nomor telepon sudah digunakan", nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(body.Password)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengenkripsi password", err.Error())
	}

	fmt.Print("Hashed Password: ", hashedPassword)

	// Inisialisasi user
	user := &models.User{

		RoleType:   models.RoleTypeSiswa, // default role
		IsVerified: true,                 // karena admin yang buat, dianggap langsung verified
	}

	// Salin data user
	if err := copier.CopyWithOption(&user, &body, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mapping data user", err.Error())
	}

	user.Password = hashedPassword

	// Inisialisasi & salin profile
	profile := &models.Profile{
		UserID: user.ID,
	}
	if err := copier.CopyWithOption(&profile, &body, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mapping data profil", err.Error())
	}

	user.Profile = profile

	// Simpan ke DB
	if err := ctrl.userService.SaveUser(user); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan user baru", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal commit transaksi", err.Error())
	}

	// Ambil ulang user
	fullUser, err := ctrl.userService.GetUserByID(user.ID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data user", err.Error())
	}

	// Mapping ke response
	var userResp dto.UserResponse
	if err := copier.Copy(&userResp, fullUser); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mapping response user", err.Error())
	}

	if err := copier.Copy(&userResp.Profile, fullUser.Profile); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mapping response profil", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "User berhasil dibuat", userResp)
}

// UpdateUserWithProfile untuk controler update user
func (ctrl *UserController) UpdateUserWithProfile(c *fiber.Ctx) error {
	// Ambil ID dari parameter
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	body := c.Locals("body").(*dto.UpdateUserWithProfileRequest)

	user, err := ctrl.userService.GetUserByID(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", err.Error())
	}

	// Copy data
	if err := copier.CopyWithOption(&user, &body, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyalin data user", err.Error())
	}

	if err := copier.CopyWithOption(&user.Profile, &body, copier.Option{IgnoreEmpty: true}); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyalin data profile", err.Error())
	}

	// Simpan ke database
	if err := ctrl.userService.SaveUser(user); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan data user", err.Error())
	}

	var usersResponse dto.UserResponse
	if copyErr := copier.Copy(&usersResponse, user); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map user data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil diperbarui", usersResponse)
}

// DeleteUserByID for delete user controller
func (ctrl *UserController) DeleteUserByID(c *fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID user tidak valid", nil)
	}

	if err := ctrl.userService.DeleteUser(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus user", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "User berhasil dihapus", nil)
}
