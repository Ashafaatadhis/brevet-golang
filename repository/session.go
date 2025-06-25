package repository

import (
	"brevet-api/models"

	"gorm.io/gorm"
)

// UserSessionRepository is a struct that represents a user session repository
type UserSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository creates a new user session repository
func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

// GetByRefreshToken retrieves a user session by its refresh token
func (r *UserSessionRepository) GetByRefreshToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.Where("refresh_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update retrieves a user session by its ID
func (r *UserSessionRepository) Update(session *models.UserSession) error {
	return r.db.Save(session).Error
}
