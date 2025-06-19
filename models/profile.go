package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Profile is a struct that represents a profile
type Profile struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null"`

	NIM      sql.NullString // NIM
	NIMProof sql.NullString // Bukti NIM
	NIK      sql.NullString // NIK

	Institution string       `gorm:"type:varchar(255)"`
	Origin      string       `gorm:"type:varchar(255)"`
	BirthDate   sql.NullTime `gorm:"type:date"`
	Address     string       `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User *User `gorm:"foreignKey:UserID"`
}
