package controllers

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/helpers"
	"brevet-api/services"
	"brevet-api/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthController represents the authentication controller
type AuthController struct {
	authService         services.IAuthService
	verificationService services.IVerificationService
	db                  *gorm.DB
}

// NewAuthController creates a new AuthController
func NewAuthController(authService services.IAuthService, verificationService services.IVerificationService, db *gorm.DB) *AuthController {
	return &AuthController{authService: authService, verificationService: verificationService, db: db}
}

// Register handles user registration
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)

	body := c.Locals("body").(*dto.RegisterRequest)
	tx := ctrl.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.WithField("recover", r).Error("Panic saat registrasi, transaksi di-rollback")
		}
	}()

	response, err := ctrl.authService.Register(ctx, tx, body)
	if err != nil {
		tx.Rollback()
		log.WithError(err).Warn("Gagal registrasi user")
		return utils.ErrorResponse(c, 400, "Gagal registrasi", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		log.WithError(err).Error("Gagal commit transaksi setelah registrasi")
		return utils.ErrorResponse(c, 500, "Gagal commit transaksi", err.Error())
	}

	log.WithField("email", body.Email).Info("Registrasi berhasil")
	return utils.SuccessResponse(c, 201, "Sukses Registrasi - Mohon cek email Anda", response)
}

// Login handles user authentication
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("Login handler called")

	body := c.Locals("body").(*dto.LoginRequest)

	result, err := ctrl.authService.Login(ctx, body, c)
	if err != nil {
		log.WithError(err).Warn("Login gagal")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Login gagal", err.Error())
	}

	// Set refresh token sebagai cookie
	env := config.GetEnv("APP_ENV", "development")
	isSecure := env == "production"

	ttlStr := config.GetEnv("REFRESH_TOKEN_EXPIRY_HOURS", "24")
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		log.WithField("ttl_raw", ttlStr).Warn("TTL tidak valid, fallback ke 24 jam")
		ttl = 24
	}

	log.WithField("user_id", result.User.ID).Info("Login sukses")
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
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("VerifyCode handler called")
	body := c.Locals("body").(*dto.VerifyRequest)

	err := ctrl.authService.VerifyUserEmail(ctx, body.Token, body.Code)
	if err != nil {
		log.WithError(err).Warn("Verifikasi email gagal")
		return utils.ErrorResponse(c, 400, "Verifikasi gagal", err.Error())
	}
	log.Info("Verifikasi email berhasil")
	return utils.SuccessResponse(c, 200, "Email verified successfully", nil)
}

// ResendVerification handles resending the verification code
func (ctrl *AuthController) ResendVerification(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	body := c.Locals("body").(*dto.ResendVerificationRequest)
	log.Info("ResendVerification handler called")
	err := ctrl.authService.ResendVerificationCode(ctx, body.Token)
	if err != nil {
		log.WithError(err).Warn("Gagal kirim ulang kode verifikasi")
		return utils.ErrorResponse(c, 400, "Gagal kirim ulang kode verifikasi", err.Error())
	}
	log.Info("Kode verifikasi berhasil dikirim ulang")
	return utils.SuccessResponse(c, 200, "Verification code resent successfully", nil)
}

// RefreshToken handles token refresh
func (ctrl *AuthController) RefreshToken(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)
	log.Info("RefreshToken handler called")
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		log.Warn("Refresh token tidak ditemukan di cookie")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Refresh token missing", nil)
	}

	tokens, err := ctrl.authService.RefreshTokens(ctx, refreshToken)
	if err != nil {
		log.WithError(err).Warn("Refresh token tidak valid atau expired")
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired refresh token", err.Error())
	}
	log.Info("Token berhasil diperbarui")
	return utils.SuccessResponse(c, fiber.StatusOK, "Token refreshed", fiber.Map{
		"access_token": tokens.AccessToken,
	})
}

// Logout handles user logout
func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()
	log := helpers.LoggerFromCtx(ctx)

	user := c.Locals("user").(*utils.Claims)

	log = log.WithField("user_id", user.UserID)
	log.Info("Logout handler called")
	refreshToken := c.Cookies("refresh_token")
	accessToken := c.Locals("access_token").(string)

	if refreshToken == "" {
		log.Warn("Refresh token tidak ditemukan saat logout")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Refresh token missing", nil)
	}

	if err := ctrl.authService.LogoutUser(ctx, accessToken, refreshToken); err != nil {
		log.WithError(err).Error("Gagal logout user")
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
	log.Info("Logout berhasil")
	return utils.SuccessResponse(c, fiber.StatusOK, "Logout successful", nil)
}
