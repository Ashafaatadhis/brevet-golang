package models

import "github.com/google/uuid"

// Role is a struct that represents a role
type Role struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string    `gorm:"size:100;not null"`
}
