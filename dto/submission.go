package dto

import (
	"time"

	"github.com/google/uuid"
)

// SubmissionResponse for response
type SubmissionResponse struct {
	ID           uuid.UUID `json:"id"`
	AssignmentID uuid.UUID `json:"assignment_id"`

	UserID uuid.UUID `json:"user_id"`

	Note      *string `json:"note"`
	EssayText *string `json:"essay_text"`

	IsLate bool `json:"is_late"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	SubmissionFiles []struct {
		ID                     uuid.UUID `json:"id"`
		AssignmentSubmissionID uuid.UUID `json:"assignment_submission_id"`
		FileURL                string    `json:"file_url"`
		CreatedAt              time.Time `json:"created_at"`
		UpdatedAt              time.Time `json:"updated_at"`
	} `json:"submission_files"`

	AssignmentGrade *SubmissionGradeResponse `json:"assignment_grade"`

	Assignment *AssignmentResponse `json:"assignment"` // Relasi ke Assignment
	User       *UserResponse       `json:"user"`       // Relasi ke User
}

// SubmissionGradeResponse represents the response structure for submission grade
type SubmissionGradeResponse struct {
	ID                     uuid.UUID `json:"id"`
	AssignmentSubmissionID uuid.UUID `json:"assignment_submission_id"` // Foreign key ke assignment_submissions.id
	Grade                  int       `json:"grade"`
	Feedback               string    `json:"feedback"`
	GradedBy               uuid.UUID `json:"graded_by"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`

	// AssignmentSubmission AssignmentSubmission `gorm:"foreignKey:AssignmentSubmissionID"`
	GradedByUser UserResponse `json:"graded_by_user"`
}

// SubmissionFileDTO submission files
type SubmissionFileDTO struct {
	FileURL string `json:"file_url"`
}

// CreateSubmissionRequest for POST
type CreateSubmissionRequest struct {
	Note            *string             `json:"note" validate:"omitempty"`
	EssayText       *string             `json:"essay_text" validate:"omitempty"`
	SubmissionFiles []SubmissionFileDTO `json:"submission_files" validate:"omitempty,dive"`
}

// UpdateSubmissionRequest for PUT
type UpdateSubmissionRequest struct {
	Note            *string              `json:"note" validate:"omitempty"`
	EssayText       *string              `json:"essay_text" validate:"omitempty"`
	SubmissionFiles *[]SubmissionFileDTO `json:"submission_files,omitempty" validate:"omitempty,dive"`
}

// GradeSubmissionRequest for grade submission request
type GradeSubmissionRequest struct {
	Grade    int    `json:"grade" validate:"required,min=0,max=100"`
	Feedback string `json:"feedback"`
}
