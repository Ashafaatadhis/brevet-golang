package utils

import (
	"brevet-api/models"
	"time"

	"gorm.io/gorm"
)

// CleanExpiredSessions is a utility function to clean up expired or revoked user sessions
func CleanExpiredSessions(db *gorm.DB) error {
	now := time.Now()
	result := db.Where("expires_at <= ? OR is_revoked = ?", now, true).Delete(&models.UserSession{})
	return result.Error
}
