package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// AssignmentResponse represents the response structure for assignments
type AssignmentResponse struct {
	ID          uuid.UUID             `json:"id"`
	MeetingID   uuid.UUID             `json:"meeting_id"`
	TeacherID   *uuid.UUID            `json:"teacher_id"`
	Title       string                `json:"title"`
	Description *string               `json:"description"`
	Type        models.AssignmentType `json:"type"`
	StartAt     time.Time             `json:"start_at"`
	EndAt       time.Time             `json:"end_at"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`

	// Meeting         models.Meeting           `json:"meeting"`
	// Teacher         *models.User             `json:"teacher"`
	AssignmentFiles []struct {
		ID           uuid.UUID `json:"id"`
		AssignmentID uuid.UUID `json:"assignment_id"`
		FileURL      string    `json:"file_url"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	} `json:"assignment_files"`
}

// AssignmentFileRequest represents the request structure for assignment files
type AssignmentFileRequest struct {
	FileURL string `json:"file_url" validate:"required"` // validasi url atau boleh pakai custom rule
}

// CreateAssignmentRequest represents the request structure for creating an assignment
type CreateAssignmentRequest struct {
	Title           string                `json:"title" validate:"required"`
	Description     *string               `json:"description" validate:"omitempty"`
	Type            models.AssignmentType `json:"type" validate:"required,assignment_type"`
	StartAt         time.Time             `json:"start_at" validate:"required"`
	EndAt           time.Time             `json:"end_at" validate:"required"`
	AssignmentFiles []string              `json:"assignment_files" validate:"omitempty,min=1,dive,required"`
}

// UpdateAssignmentRequest represents the request structure for updating an assignment
type UpdateAssignmentRequest struct {
	Title           *string                `json:"title" validate:"omitempty"`
	Description     *string                `json:"description" validate:"omitempty"`
	Type            *models.AssignmentType `json:"type" validate:"omitempty,assignment_type"`
	StartAt         *time.Time             `json:"start_at" validate:"omitempty"`
	EndAt           *time.Time             `json:"end_at" validate:"omitempty"`
	AssignmentFiles []string               `json:"assignment_files" validate:"omitempty,dive,required"`
}
