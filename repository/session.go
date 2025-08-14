package repository

import (
	"brevet-api/models"
	"context"

	"gorm.io/gorm"
)

// IUserSessionRepository interface
type IUserSessionRepository interface {
	GetByRefreshToken(ctx context.Context, token string) (*models.UserSession, error)
	Update(ctx context.Context, session *models.UserSession) error
}

// UserSessionRepository is a struct that represents a user session repository
type UserSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository creates a new user session repository
func NewUserSessionRepository(db *gorm.DB) IUserSessionRepository {
	return &UserSessionRepository{db: db}
}

// GetByRefreshToken retrieves a user session by its refresh token
func (r *UserSessionRepository) GetByRefreshToken(ctx context.Context, token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).Where("refresh_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update retrieves a user session by its ID
func (r *UserSessionRepository) Update(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}
