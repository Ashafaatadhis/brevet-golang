package models

import (
	"time"

	"github.com/google/uuid"
)

// GroupDaysBatch is a struct that represents a group days batch
type GroupDaysBatch struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID uuid.UUID `gorm:"type:uuid;not null"`
	Batch   Batch     `gorm:"foreignKey:BatchID;references:ID"`
	DayID   uuid.UUID `gorm:"type:uuid;not null"`
	Day     Day       `gorm:"foreignKey:DayID;references:ID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
