package models

import (
	"time"

	"github.com/google/uuid"
)

// GroupDaysBatch is a struct that represents a group days batch
type GroupDaysBatch struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID   uuid.UUID `gorm:"type:uuid;not null"`
	Batch     Batch     `gorm:"foreignKey:BatchID;references:ID"`
	Day       DayType   `gorm:"type:day_type;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
