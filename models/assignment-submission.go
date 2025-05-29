package models

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentSubmission represents the assignment_submissions table
type AssignmentSubmission struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AssignmentID uuid.UUID `gorm:"type:uuid;not null"`

	UserID uuid.UUID `gorm:"type:uuid;not null"`

	Note        string    `gorm:"type:text"`
	SubmittedAt time.Time `gorm:"type:timestamp"`
	IsLate      bool      `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Assignment Assignment `gorm:"foreignKey:AssignmentID;references:ID"` // Relasi ke Assignment
	User       User       `gorm:"foreignKey:UserID;references:ID"`       // Relasi ke User
}
