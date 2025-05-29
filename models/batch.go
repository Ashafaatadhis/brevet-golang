package models

import (
	"time"

	"github.com/google/uuid"
)

// Batch is a struct that represents a batch
type Batch struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CourseID     uuid.UUID `gorm:"type:uuid;not null"`
	Title        string    `gorm:"type:varchar(255);not null"`
	Description  string    `gorm:"type:text"`
	CourseTypeID uuid.UUID `gorm:"type:uuid;not null"`

	BatchThumbnail string    `gorm:"type:varchar(255)"`
	StartAt        time.Time `gorm:"type:timestamp"`
	EndAt          time.Time `gorm:"type:timestamp"`
	Room           string    `gorm:"type:varchar(255)"`
	Quota          int       `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	Course     Course     `gorm:"foreignKey:CourseID"`
	CourseType CourseType `gorm:"foreignKey:CourseTypeID"`
}
