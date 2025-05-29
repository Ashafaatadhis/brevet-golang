package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User is a struct that represents a user
type User struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name       string    `gorm:"size:100;not null"`
	Phone      string    `gorm:"size:20;unique;not null"`
	Avatar     string    `gorm:"size:255"`
	RoleID     uuid.UUID `gorm:"type:uuid;not null"`
	Email      string    `gorm:"size:100;unique;not null"`
	Password   string    `gorm:"size:255;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	IsVerified bool `gorm:"default:false"`
	VerifyCode sql.NullString
	CodeExpiry sql.NullTime
	LastSentAt sql.NullTime `gorm:"default:NULL"`

	// Foreign Key
	Profile *Profile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Role    Role     `gorm:"foreignKey:RoleID"`
}
