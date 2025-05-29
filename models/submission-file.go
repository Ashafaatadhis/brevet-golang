package models

import (
	"time"

	"github.com/google/uuid"
)

// SubmissionFile represents the submission_files table
type SubmissionFile struct {
	ID                     uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AssignmentSubmissionID uuid.UUID `gorm:"type:uuid;not null"` // Foreign key ke assignment_submissions.id
	FileURL                string    `gorm:"type:varchar(255);not null"`
	CreatedAt              time.Time
	UpdatedAt              time.Time

	AssignmentSubmission AssignmentSubmission `gorm:"foreignKey:AssignmentSubmissionID"`
}
