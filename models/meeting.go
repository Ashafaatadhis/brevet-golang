package models

import (
	"time"

	"github.com/google/uuid"
)

// Meeting is a struct that represents a meeting
type Meeting struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID uuid.UUID `gorm:"type:uuid;not null"`
	// TeacherID   uuid.UUID   `gorm:"type:uuid;not null"` // <--- Tambahkan ini
	Title       string      `gorm:"type:varchar(255);not null"`
	Description string      `gorm:"type:text"`
	Type        MeetingType `gorm:"type:meeting_type;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Batch       Batch        `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`
	Quizzes     []Quiz       `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	Assignments []Assignment `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	Materials   []Material   `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	Teachers    []User       `gorm:"many2many:meeting_teachers;constraint:OnDelete:CASCADE"`
}
