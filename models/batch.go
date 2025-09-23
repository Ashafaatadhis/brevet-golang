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
	StartAt        time.Time `gorm:"type:timestamptz;not null"`
	EndAt          time.Time `gorm:"type:timestamptz;not null"`
	// StartTime      time.Time `gorm:"type:time;not null"`
	// EndTime        time.Time `gorm:"type:time;not null"`
	StartTime string `gorm:"type:time without time zone;default:'08:00:00';not null"`
	EndTime   string `gorm:"type:time without time zone;default:'10:00:00';not null"`

	// Rentang pendaftaran
	RegistrationStartAt time.Time `gorm:"type:timestamptz;not null;default:now()"`
	RegistrationEndAt   time.Time `gorm:"type:timestamptz;not null;default:now()"`

	Room      string    `gorm:"type:varchar(255);not null"`
	Quota     int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	BatchGroups []BatchGroup `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`

	BatchDays  []BatchDay `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`
	Course     Course     `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	CourseType CourseType `gorm:"type:course_type;not null"`
}
