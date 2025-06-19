package models

import (
	"time"

	"github.com/google/uuid"
)

// Price is a struct that represents a price
type Price struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	GroupType GroupType `gorm:"type:group_type;not null"`

	Price float64 `gorm:"type:numeric;not null"` // Harga

	CreatedAt time.Time
	UpdatedAt time.Time
}
