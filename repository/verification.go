package repository

import (
	"brevet-api/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IVerificationRepository interface
type IVerificationRepository interface {
	WithTx(tx *gorm.DB) IVerificationRepository
	UpdateVerificationCode(ctx context.Context, userID uuid.UUID, code string, expiry time.Time) error
	FindUserByCode(ctx context.Context, userID uuid.UUID, code string) (*models.User, error)
	MarkUserVerified(ctx context.Context, userID uuid.UUID) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

// VerificationRepository is a struct that represents a verification repository
type VerificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *gorm.DB) IVerificationRepository {
	return &VerificationRepository{db: db}
}

// WithTx running with transaction
func (r *VerificationRepository) WithTx(tx *gorm.DB) IVerificationRepository {
	return &VerificationRepository{db: tx}
}

// UpdateVerificationCode updates the verification code and expiry for a user
func (r *VerificationRepository) UpdateVerificationCode(ctx context.Context, userID uuid.UUID, code string, expiry time.Time) error {

	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"verify_code":  code,
			"code_expiry":  expiry,
			"last_sent_at": time.Now(),
		}).Error
}

// FindUserByCode finds a user by their verification code and checks if the code is still valid
func (r *VerificationRepository) FindUserByCode(ctx context.Context, userID uuid.UUID, code string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ? AND verify_code = ? AND code_expiry > ?", userID, code, time.Now()).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// MarkUserVerified marks a user as verified by clearing their verification code and expiry
func (r *VerificationRepository) MarkUserVerified(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"is_verified": true,
			"verify_code": nil,
			"code_expiry": nil,
		}).Error
}

// GetUserByID retrieves a user by their ID
func (r *VerificationRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
