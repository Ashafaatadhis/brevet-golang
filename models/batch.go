package models

import (
	"time"

	"github.com/google/uuid"
)

// Batch is a struct that represents a batch
type Batch struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Slug           string    `gorm:"type:varchar(255);not null;unique"`
	CourseID       uuid.UUID `gorm:"type:uuid;not null"`
	Title          string    `gorm:"type:varchar(255);not null"`
	Description    string    `gorm:"type:text"`
	BatchThumbnail string    `gorm:"type:varchar(255)"`
	StartAt        time.Time `gorm:"type:timestamp;not null"`
	EndAt          time.Time `gorm:"type:timestamp;not null"`
	Room           string    `gorm:"type:varchar(255);not null"`
	Quota          int       `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	Course     Course     `gorm:"foreignKey:CourseID"`
	CourseType CourseType `gorm:"type:course_type;not null"`
}
