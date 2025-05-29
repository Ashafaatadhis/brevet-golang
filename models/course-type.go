package models

import "github.com/google/uuid"

// CourseType is a struct that represents a course type
type CourseType struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name string    `gorm:"size:100;not null"`
}
