package models

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentFiles is a struct that represents an assignment file
type AssignmentFiles struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AssignmentID uuid.UUID `gorm:"type:uuid;not null"`
	FileURL      string    `gorm:"type:text;not null"` // gunakan type:text untuk URL panjang
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Assignment Assignment `gorm:"foreignKey:AssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
