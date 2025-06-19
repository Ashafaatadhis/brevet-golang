package services

import (
	"brevet-api/config"
	"brevet-api/models"
	"brevet-api/utils"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService is a struct that represents a user service
type AuthService struct {
	db *gorm.DB
}

// NewAuthService creates a new user service
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// IsEmailUnique checks if email is unique
func (s *AuthService) IsEmailUnique(db *gorm.DB, email string) bool {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsPhoneUnique checks if phone is unique
func (s *AuthService) IsPhoneUnique(db *gorm.DB, phone string) bool {
	var user models.User
	err := db.Where("phone = ?", phone).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// CreateUser creates a new user in database
func (s *AuthService) CreateUser(db *gorm.DB, user *models.User) error {
	if err := db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// CreateProfile creates a new profile in database
func (s *AuthService) CreateProfile(db *gorm.DB, profile *models.Profile) error {
	if err := db.Create(profile).Error; err != nil {
		return err
	}
	return nil
}

// GetUsers gets user
func (s *AuthService) GetUsers(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds user by email with role information
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailWithProfile finds user by email with role and profile information
func (s *AuthService) GetUserByEmailWithProfile(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Profile").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByIDWithProfile is
func (s *AuthService) GetUserByIDWithProfile(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID gets a user by their ID
func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// TokenPair represents a pair of tokens
type TokenPair struct {
	AccessToken string
}

// RefreshTokens refreshes tokens
func (s *AuthService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	refreshTokenSecret := config.GetEnv("REFRESH_TOKEN_SECRET", "default-key")
	accessTokenSecret := config.GetEnv("ACCESS_TOKEN_SECRET", "default-key")

	// 1. Validasi token JWT
	claims, err := utils.ExtractClaimsFromToken(refreshToken, refreshTokenSecret)
	if err != nil || claims == nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// 2. Cek session refresh token di DB
	var session models.UserSession
	if err = s.db.Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		return nil, fmt.Errorf("refresh token session not found")
	}
	if session.IsRevoked {
		return nil, fmt.Errorf("refresh token revoked")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}

	// 3. Ambil data user
	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// 4. Generate access token baru
	accessTokenExpiryStr := config.GetEnv("ACCESS_TOKEN_EXPIRY_HOURS", "24")
	accessTokenExpiryHours, err := strconv.Atoi(accessTokenExpiryStr)
	if err != nil {
		accessTokenExpiryHours = 24
	}
	accessToken, err := utils.GenerateJWT(*user, accessTokenSecret, accessTokenExpiryHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 5. (Optional) generate refresh token baru juga, simpan & revoke session lama
	// Kalau mau implement rotate refresh token, bisa dilakukan di sini.

	return &TokenPair{
		AccessToken: accessToken,
		// RefreshToken: newRefreshToken, // kalau generate baru
	}, nil
}

// CreateUserSession creates a new user session
func (s *AuthService) CreateUserSession(userID uuid.UUID, refreshToken string, c *fiber.Ctx) error {
	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	refreshTokenExpiryStr := config.GetEnv("REFRESH_TOKEN_EXPIRY_HOURS", "24")
	refreshTokenExpiryHours, err := strconv.Atoi(refreshTokenExpiryStr)
	if err != nil {
		refreshTokenExpiryHours = 24
	}

	expiresAt := time.Now().Add(time.Duration(refreshTokenExpiryHours) * time.Hour)

	session := models.UserSession{
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    sql.NullString{String: userAgent, Valid: userAgent != ""},
		IPAddress:    sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		ExpiresAt:    expiresAt,
		IsRevoked:    false,
	}

	return s.db.Create(&session).Error
}

// RevokeUserSessionByRefreshToken revokes a user session by refresh token
func (s *AuthService) RevokeUserSessionByRefreshToken(refreshToken string) error {
	var session models.UserSession
	if err := s.db.Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		return fmt.Errorf("refresh token session not found")
	}

	// Update session jadi revoked
	session.IsRevoked = true
	if err := s.db.Save(&session).Error; err != nil {
		return err
	}

	return nil
}
