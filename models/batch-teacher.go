package models

import (
	"time"

	"github.com/google/uuid"
)

// BatchTeacher is a struct that represents a batch teacher
type BatchTeacher struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID  uuid.UUID `gorm:"type:uuid;not null"`
	BatchID uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Batch Batch `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`
}
