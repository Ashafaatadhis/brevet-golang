package repository

import (
	"brevet-api/config"
	"brevet-api/models"
	"context"

	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAuthRepository interface
type IAuthRepository interface {
	WithTx(tx *gorm.DB) IAuthRepository
	IsEmailUnique(ctx context.Context, email string) bool
	IsPhoneUnique(ctx context.Context, phone string) bool
	CreateUser(ctx context.Context, user *models.User) error
	CreateProfile(ctx context.Context, profile *models.Profile) error
	GetUsers(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByEmailWithProfile(ctx context.Context, email string) (*models.User, error)
	GetUserByIDWithProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	CreateUserSession(ctx context.Context, userID uuid.UUID, refreshToken string, c *fiber.Ctx) error
	RevokeUserSessionByRefreshToken(ctx context.Context, refreshToken string) error
}

// AuthRepository is a struct that represents a user service
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new user repository
func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{db: db}
}

// WithTx running with transaction
func (s *AuthRepository) WithTx(tx *gorm.DB) IAuthRepository {
	return &AuthRepository{db: tx}
}

// IsEmailUnique checks if email is unique
func (s *AuthRepository) IsEmailUnique(ctx context.Context, email string) bool {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsPhoneUnique checks if phone is unique
func (s *AuthRepository) IsPhoneUnique(ctx context.Context, phone string) bool {
	var user models.User
	err := s.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// CreateUser creates a new user in database
func (s *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

// CreateProfile creates a new profile in database
func (s *AuthRepository) CreateProfile(ctx context.Context, profile *models.Profile) error {
	if err := s.db.WithContext(ctx).Create(profile).Error; err != nil {
		return err
	}
	return nil
}

// GetUsers gets user
func (s *AuthRepository) GetUsers(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail finds user by email with role information
func (s *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailWithProfile finds user by email with role and profile information
func (s *AuthRepository) GetUserByEmailWithProfile(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Profile").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByIDWithProfile is
func (s *AuthRepository) GetUserByIDWithProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID gets a user by their ID
func (s *AuthRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// // GetUserByIDTx gets a user by their ID within a transaction
// func (s *AuthRepository) GetUserByIDTx(tx *gorm.DB, userID uuid.UUID) (*models.User, error) {
// 	var user models.User
// 	if err := tx.Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
// 		return nil, err
// 	}
// 	return &user, nil
// }

// CreateUserSession creates a new user session
func (s *AuthRepository) CreateUserSession(ctx context.Context, userID uuid.UUID, refreshToken string, c *fiber.Ctx) error {

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

	return s.db.WithContext(ctx).Create(&session).Error
}

// RevokeUserSessionByRefreshToken revokes a user session by refresh token
func (s *AuthRepository) RevokeUserSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	var session models.UserSession
	if err := s.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&session).Error; err != nil {
		return fmt.Errorf("refresh token session not found")
	}

	// Update session jadi revoked
	session.IsRevoked = true
	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		return err
	}

	return nil
}
