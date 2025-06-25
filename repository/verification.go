package repository

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationRepository is a struct that represents a verification repository
type VerificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *gorm.DB) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// UpdateVerificationCode updates the verification code and expiry for a user
func (r *VerificationRepository) UpdateVerificationCode(tx *gorm.DB, userID uuid.UUID, code string, expiry time.Time) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"verify_code":  code,
			"code_expiry":  expiry,
			"last_sent_at": time.Now(),
		}).Error
}

// FindUserByCode finds a user by their verification code and checks if the code is still valid
func (r *VerificationRepository) FindUserByCode(userID uuid.UUID, code string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND verify_code = ? AND code_expiry > ?", userID, code, time.Now()).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// MarkUserVerified marks a user as verified by clearing their verification code and expiry
func (r *VerificationRepository) MarkUserVerified(userID uuid.UUID) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"is_verified": true,
			"verify_code": nil,
			"code_expiry": nil,
		}).Error
}

// GetUserByID retrieves a user by their ID
func (r *VerificationRepository) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
