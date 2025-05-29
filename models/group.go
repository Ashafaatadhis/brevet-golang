package models

import "github.com/google/uuid"

// Group is a struct that represents a group
type Group struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string    `gorm:"size:100;not null"`
}
