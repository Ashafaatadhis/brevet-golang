package models

import (
	"time"

	"github.com/google/uuid"
)

// Meeting is a struct that represents a meeting
type Meeting struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID     uuid.UUID `gorm:"type:uuid;not null"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Batch Batch `gorm:"foreignKey:BatchID"`
}
