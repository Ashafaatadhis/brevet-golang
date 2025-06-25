package repository

import (
	"brevet-api/config"
	"brevet-api/models"

	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthRepository is a struct that represents a user service
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new user repository
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// IsEmailUnique checks if email is unique
func (s *AuthRepository) IsEmailUnique(db *gorm.DB, email string) bool {
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsPhoneUnique checks if phone is unique
func (s *AuthRepository) IsPhoneUnique(db *gorm.DB, phone string) bool {
	var user models.User
	err := db.Where("phone = ?", phone).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// CreateUser creates a new user in database
func (s *AuthRepository) CreateUser(db *gorm.DB, user *models.User) error {
	if err := db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// CreateProfile creates a new profile in database
func (s *AuthRepository) CreateProfile(db *gorm.DB, profile *models.Profile) error {
	if err := db.Create(profile).Error; err != nil {
		return err
	}
	return nil
}

// GetUsers gets user
func (s *AuthRepository) GetUsers(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds user by email with role information
func (s *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailWithProfile finds user by email with role and profile information
func (s *AuthRepository) GetUserByEmailWithProfile(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Profile").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByIDWithProfile is
func (s *AuthRepository) GetUserByIDWithProfile(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID gets a user by their ID
func (s *AuthRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	fmt.Println("User found:", user)

	return &user, nil
}

// GetUserByIDTx gets a user by their ID within a transaction
func (s *AuthRepository) GetUserByIDTx(tx *gorm.DB, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := tx.Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUserSession creates a new user session
func (s *AuthRepository) CreateUserSession(userID uuid.UUID, refreshToken string, c *fiber.Ctx) error {
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
func (s *AuthRepository) RevokeUserSessionByRefreshToken(refreshToken string) error {
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
