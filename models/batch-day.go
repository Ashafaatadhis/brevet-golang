package models

import "github.com/google/uuid"

// BatchDay pivot table
type BatchDay struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID uuid.UUID `gorm:"type:uuid;not null"`
	Day     DayType   `gorm:"type:day_type;not null"`

	Batch Batch `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`
}
