package models

import (
	"time"

	"github.com/google/uuid"
)

// Price is a struct that represents a price
type Price struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GroupID uuid.UUID `gorm:"type:uuid;not null"`               // Foreign key ke groups.id
	Group   Group     `gorm:"foreignKey:GroupID;references:ID"` // Relasi ke Group

	Price float64 `gorm:"type:numeric;not null"` // Harga

	CreatedAt time.Time
	UpdatedAt time.Time
}
