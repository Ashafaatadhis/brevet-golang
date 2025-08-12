package models

import (
	"time"

	"github.com/google/uuid"
)

// AssignmentGrade represents the assignment_grades table
type AssignmentGrade struct {
	ID                     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AssignmentSubmissionID uuid.UUID `gorm:"type:uuid;not null"` // Foreign key ke assignment_submissions.id
	Grade                  int       `gorm:"not null"`
	Feedback               string    `gorm:"type:text"`
	GradedBy               uuid.UUID `gorm:"type:uuid;not null"` // Foreign key ke users.id
	// GradedAt               time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	AssignmentSubmission AssignmentSubmission `gorm:"foreignKey:AssignmentSubmissionID"`
	GradedByUser         User                 `gorm:"foreignKey:GradedBy"`
}
