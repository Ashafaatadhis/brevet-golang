package models

import (
	"time"

	"github.com/google/uuid"
)

// CourseImage represents the course image data model
type CourseImage struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CourseID  uuid.UUID `gorm:"type:uuid;not null"`
	ImageURL  string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Course Course `gorm:"foreignKey:CourseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
