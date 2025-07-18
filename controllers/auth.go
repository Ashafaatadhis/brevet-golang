package controllers

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthController represents the authentication controller
type AuthController struct {
	authService *services.AuthService

	verificationService *services.VerificationService
	db                  *gorm.DB // ← tambahkan ini
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *services.AuthService, verificationService *services.VerificationService, db *gorm.DB) *AuthController {
	return &AuthController{authService: authService, verificationService: verificationService, db: db}
}

// Register handles user registration
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.RegisterRequest)
	tx := ctrl.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	response, err := ctrl.authService.Register(tx, body)
	if err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, 400, "Gagal registrasi", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, 500, "Gagal commit transaksi", err.Error())
	}

	return utils.SuccessResponse(c, 201, "Sukses Registrasi - Mohon cek email Anda", response)
}

// Login handles user authentication
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.LoginRequest)

	result, err := ctrl.authService.Login(body, c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Login gagal", err.Error())
	}

	// Set refresh token sebagai cookie
	env := config.GetEnv("APP_ENV", "development")
	isSecure := env == "production"

	ttlStr := config.GetEnv("REFRESH_TOKEN_EXPIRY_HOURS", "24")
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 24
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "None",
		Expires:  time.Now().Add(time.Duration(ttl) * time.Hour),
		Path:     "/",
	})

	return utils.SuccessResponse(c, 200, "Login successful", fiber.Map{
		"access_token": result.AccessToken,
		"user":         result.User,
	})
}

// VerifyCode handles email verification
func (ctrl *AuthController) VerifyCode(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.VerifyRequest)

	err := ctrl.authService.VerifyUserEmail(body.Token, body.Code)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Verifikasi gagal", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Email verified successfully", nil)
}

// ResendVerification handles resending the verification code
func (ctrl *AuthController) ResendVerification(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.ResendVerificationRequest)

	err := ctrl.authService.ResendVerificationCode(body.Token)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Gagal kirim ulang kode verifikasi", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Verification code resent successfully", nil)
}

// RefreshToken handles token refresh
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Refresh token missing", nil)
	}

	tokens, err := ctrl.authService.RefreshTokens(refreshToken)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired refresh token", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Token refreshed", fiber.Map{
		"access_token": tokens.AccessToken,
	})
}

// Logout handles user logout
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	accessToken := c.Locals("access_token").(string)

	if refreshToken == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Refresh token missing", nil)
	}

	if err := ctrl.authService.LogoutUser(accessToken, refreshToken); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to logout", err.Error())
	}

	// Hapus cookie
	env := config.GetEnv("APP_ENV", "development")
	isSecure := env == "production"

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "None",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Logout successful", nil)
}
