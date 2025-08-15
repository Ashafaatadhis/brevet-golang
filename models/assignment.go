package models

import (
	"time"

	"github.com/google/uuid"
)

// Assignment is a struct that represents an assignment
type Assignment struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MeetingID   uuid.UUID      `gorm:"type:uuid;not null"`
	TeacherID   uuid.UUID      `gorm:"type:uuid;not null"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Description *string        `gorm:"type:text"`
	Type        AssignmentType `gorm:"type:assignment_type;not null"`
	StartAt     time.Time      `gorm:"type:timestamptz"`
	EndAt       time.Time      `gorm:"type:timestamptz"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Meeting         Meeting           `gorm:"foreignKey:MeetingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Teacher         *User             `gorm:"foreignKey:TeacherID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AssignmentFiles []AssignmentFiles `gorm:"foreignKey:AssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	AssignmentSubmissions []AssignmentSubmission `gorm:"foreignKey:AssignmentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
