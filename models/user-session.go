package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// UserSession is a model for user sessions
type UserSession struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	User         User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // relasi
	RefreshToken string         `gorm:"type:text;not null"`
	UserAgent    sql.NullString `gorm:"type:text"`
	IPAddress    sql.NullString `gorm:"type:varchar(45)"` // IPv4/IPv6
	IsRevoked    bool           `gorm:"default:false"`
	ExpiresAt    time.Time      `gorm:"not null;index"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
}
