package controllers

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// AuthController represents the authentication controller
type AuthController struct {
	authService         *services.AuthService
	roleService         *services.RoleService
	verificationService *services.VerificationService
	db                  *gorm.DB // ← tambahkan ini
}

// NewAuthController creates a new AuthController
func NewAuthController(authService *services.AuthService, roleService *services.RoleService, verificationService *services.VerificationService, db *gorm.DB) *AuthController {
	return &AuthController{authService: authService, roleService: roleService, verificationService: verificationService, db: db}
}

// Register handles user registration
func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.RegisterRequest)
	tx := ctrl.db.Begin() // Start a transaction

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if !ctrl.authService.IsEmailUnique(tx, body.Email) || !ctrl.authService.IsPhoneUnique(tx, body.Phone) {
		tx.Rollback()
		return utils.ErrorResponse(c, 400, "Email or phone is already registered", nil)
	}

	// Get default role
	role, roleErr := ctrl.roleService.GetRoleByName(tx, "siswa")
	if roleErr != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, 500, "Failed to find default role", roleErr.Error())
	}

	// Create user
	var user models.User
	if copyErr := copier.Copy(&user, body); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map user data", copyErr.Error())
	}

	// Hash password
	hashedPassword, hashErr := utils.HashPassword(body.Password)
	if hashErr != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, 500, "Failed to secure password", hashErr.Error())
	}

	user.Password = hashedPassword
	user.RoleID = role.ID

	if createUserErr := ctrl.authService.CreateUser(tx, &user); createUserErr != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, 500, "Failed to create user", createUserErr.Error())
	}

	// Generate verification code

	code, err := ctrl.verificationService.GenerateVerificationCode(tx, user.ID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to generate verification code", err.Error())
	}

	// generate JWT token
	token, err := utils.GenerateVerificationToken(user.ID, user.Email)
	if err != nil {

		return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
	}

	// Send verification email (implement this)
	go utils.SendVerificationEmail(user.Email, code, token)

	// Create profile
	var profile models.Profile
	if profileCopyErr := copier.Copy(&profile, body); profileCopyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map profile data", profileCopyErr.Error())
	}
	profile.UserID = user.ID

	if createProfileErr := ctrl.authService.CreateProfile(tx, &profile); createProfileErr != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, 500, "Failed to create profile", createProfileErr.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return utils.ErrorResponse(c, 500, "Gagal commit transaksi", err.Error())
	}

	// Get user with role
	fullUser, getUserErr := ctrl.authService.GetUserByIDWithRole(user.ID)
	if getUserErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to get user details", getUserErr.Error())
	}

	// Map to response
	var response dto.RegisterResponse
	if respCopyErr := copier.Copy(&response, &fullUser); respCopyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to create response", respCopyErr.Error())
	}
	if profileRespCopyErr := copier.Copy(&response.Profile, &profile); profileRespCopyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to create profile response", profileRespCopyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Registration successful - please verify your email", response)
}

// Login handles user authentication
func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.LoginRequest)

	// Find user by email with role
	user, err := ctrl.authService.GetUserByEmailWithRoleAndProfile(body.Email)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Invalid credentials", err.Error())
	}

	// Verify password
	if !utils.CheckPasswordHash(body.Password, user.Password) {
		return utils.ErrorResponse(c, 401, "Invalid credentials", nil)
	}

	// Check if user is verified
	if !user.IsVerified {
		return utils.ErrorResponse(c, 401, "Email not verified. Please verify your email first.", nil)
	}

	// Generate access token
	accessTokenSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key")
	accessTokenExpiryStr := config.GetEnv("ACCESS_TOKEN_EXPIRY_HOURS", "24")
	accessTokenExpiryHours, err := strconv.Atoi(accessTokenExpiryStr)
	if err != nil {
		accessTokenExpiryHours = 24
	}
	accessToken, err := utils.GenerateJWT(*user, accessTokenSecret, accessTokenExpiryHours)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to generate access token", err.Error())
	}

	// Generate refresh token
	refreshTokenSecret := config.GetEnv("REFRESH_TOKEN_SECRET", "default-key")
	refreshTokenExpiryStr := config.GetEnv("REFRESH_TOKEN_EXPIRY_HOURS", "24")
	refreshTokenExpiryHours, err := strconv.Atoi(refreshTokenExpiryStr)
	if err != nil {
		refreshTokenExpiryHours = 24
	}
	refreshToken, err := utils.GenerateJWT(*user, refreshTokenSecret, refreshTokenExpiryHours)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to generate refresh token", err.Error())
	}

	if err := ctrl.authService.CreateUserSession(user.ID, refreshToken, c); err != nil {
		return utils.ErrorResponse(c, 500, "Failed to save user session", err.Error())
	}

	// Set refresh token di HTTP-only cookie
	env := config.GetEnv("APP_ENV", "development")
	isSecure := false
	if env == "production" {
		isSecure = true
	}
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   isSecure,
		Path:     "/",
		SameSite: "Lax",
	})

	// Map to response
	var userResponse dto.UserResponse
	if respCopyErr := copier.Copy(&userResponse, &user); respCopyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to create response", respCopyErr.Error())
	}
	if respCopyErr := copier.Copy(&userResponse.Profile, &user.Profile); respCopyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to create response", respCopyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Login successful", fiber.Map{
		"access_token": accessToken,
		"user":         userResponse,
	})
}

// VerifyCode handles email verification
func (ctrl *AuthController) VerifyCode(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.VerifyRequest)

	// Extract user ID from JWT token
	jwtSecret := config.GetEnv("VERIFICATION_TOKEN_SECRET", "default-key") // Consider making this required in production

	payload, err := utils.ExtractUserIDFromToken(body.Token, jwtSecret)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Invalid token", err.Error())
	}

	// Ambil user
	user, err := ctrl.authService.GetUserByID(payload.UserID)
	if err != nil {
		return utils.ErrorResponse(c, 404, "User not found", err.Error())
	}

	// Sudah diverifikasi
	if user.IsVerified {
		return utils.ErrorResponse(c, 400, "Email already verified", nil)
	}

	// Verify the code
	isValid := ctrl.verificationService.VerifyCode(payload.UserID, body.Code)
	if !isValid {
		return utils.ErrorResponse(c, 400, "Invalid or expired verification code", nil)
	}

	return utils.SuccessResponse(c, 200, "Email verified successfully", nil)
}

// ResendVerification handles resending verification code
func (ctrl *AuthController) ResendVerification(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.ResendVerificationRequest)

	// Extract data dari token
	jwtSecret := config.GetEnv("VERIFICATION_TOKEN_SECRET", "default-key") // Consider making this required in production
	payload, err := utils.ExtractUserIDFromToken(body.Token, jwtSecret)
	if err != nil {
		return utils.ErrorResponse(c, 401, "Invalid token", nil)
	}

	// Ambil user dari database
	user, err := ctrl.authService.GetUserByID(payload.UserID)
	if err != nil {
		return utils.ErrorResponse(c, 404, "User not found", err.Error())
	}

	// Cek apakah sudah diverifikasi
	if user.IsVerified {
		return utils.ErrorResponse(c, 400, "Email already verified", nil)
	}

	// Check cooldown
	remaining, err := ctrl.verificationService.GetCooldownRemaining(user.ID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to check resend cooldown", err.Error())
	}
	if remaining > 0 {
		seconds := int(remaining.Seconds())
		return utils.ErrorResponse(c, 429, fmt.Sprintf("Please wait %d seconds before requesting a new code", seconds), nil)
	}

	// Generate verification code dan update ke DB
	code, err := ctrl.verificationService.GenerateVerificationCode(nil, user.ID)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to generate verification code", err.Error())
	}

	// Generate JWT token baru untuk verifikasi
	token, err := utils.GenerateVerificationToken(user.ID, user.Email)
	if err != nil {
		return utils.ErrorResponse(c, 500, "Failed to generate token", err.Error())
	}

	// Kirim email verifikasi
	go utils.SendVerificationEmail(user.Email, code, token)

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
	token := c.Locals("access_token").(string)

	if refreshToken == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Refresh token missing", nil)
	}

	// Panggil service untuk revoke session
	if err := ctrl.authService.RevokeUserSessionByRefreshToken(refreshToken); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to revoke session", err.Error())
	}

	ttlStr := config.GetEnv("TOKEN_BLACKLIST_TTL", "86400") // default 24 jam
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		ttl = 86400 // fallback 24 jam
	}

	// Set token ke Redis dengan TTL (misal 24 jam)
	err = config.RedisClient.Set(config.Ctx, token, "blacklisted", time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to blacklist token", err.Error())
	}

	// Hapus cookie refresh token di client
	env := config.GetEnv("APP_ENV", "development")
	isSecure := false
	if env == "production" {
		isSecure = true
	}
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   isSecure,
		Path:     "/",
		SameSite: "Lax",
		Expires:  time.Unix(0, 0), // Set cookie expired di masa lalu supaya terhapus
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "Logout successful", nil)
}
