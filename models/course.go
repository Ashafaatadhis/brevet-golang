package models

import (
	"time"

	"github.com/google/uuid"
)

// Course represents the course data model
type Course struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Title            string    `gorm:"type:varchar(255);not null"`
	ShortDescription string    `gorm:"type:varchar"`
	Description      string    `gorm:"type:text"`
	LearningOutcomes string    `gorm:"type:text"`
	Achievements     string    `gorm:"type:text"`

	// CourseThumbnail string `gorm:"type:varchar"`
	CourseImages []CourseImage `gorm:"foreignKey:CourseID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
