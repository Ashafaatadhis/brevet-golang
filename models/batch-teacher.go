package models

import (
	"time"

	"github.com/google/uuid"
)

// BatchTeacher is a struct that represents a batch teacher
type BatchTeacher struct {
	ID     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null"` // Foreign key ke users.id

	BatchID   uuid.UUID `gorm:"type:uuid;not null"` // Foreign key ke batches.id
	CreatedAt time.Time
	UpdatedAt time.Time

	User  User  `gorm:"foreignKey:UserID;references:ID"`
	Batch Batch `gorm:"foreignKey:BatchID;references:ID"`
}
