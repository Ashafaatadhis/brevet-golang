package models

import "github.com/google/uuid"

// Day is a struct that represents a day
type Day struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string    `gorm:"size:100;not null"`
}
