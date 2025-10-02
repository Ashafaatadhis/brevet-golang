package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// MeetingResponse is struct for response meeting
type MeetingResponse struct {
	ID          uuid.UUID          `json:"id"`
	BatchID     uuid.UUID          `json:"batch_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Type        models.MeetingType `json:"meeting_type"`

	StartAt time.Time `json:"start_at"` // default: StartAt = NOW()
	EndAt   time.Time `json:"end_at"`   // default: EndAt = NOW() + interval '2 hours'
	IsOpen  bool      `json:"is_open"`

	Teachers    []UserResponse       `json:"teachers"`
	Assignments []AssignmentResponse `json:"assignments"`
	Materials   []MaterialResponse   `json:"materials"`
	Quizzes     []QuizResponse       `json:"quizzes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMeetingRequest is request income
type CreateMeetingRequest struct {
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description"`
	Type        models.MeetingType `json:"type" validate:"required,meeting_type"`

	StartAt time.Time `json:"start_at" validate:"required"` // default: StartAt = NOW()
	EndAt   time.Time `json:"end_at" validate:"required"`   // default: EndAt = NOW() + interval '2 hours'
	IsOpen  *bool     `json:"is_open" validate:"omitempty"`
}

// UpdateMeetingRequest is request income
type UpdateMeetingRequest struct {
	Title       *string             `json:"title" validate:"omitempty"`
	Description *string             `json:"description" validate:"omitempty"`
	Type        *models.MeetingType `json:"type" validate:"omitempty,meeting_type"`
	StartAt     *time.Time          `json:"start_at" validate:"omitempty"` // default: StartAt = NOW()
	EndAt       *time.Time          `json:"end_at" validate:"omitempty"`   // default: EndAt = NOW() + interval '2 hours'
	IsOpen      *bool               `json:"is_open" validate:"omitempty"`
}

// AssignTeachersRequest AssignTeachersRequest is request income
type AssignTeachersRequest struct {
	TeacherIDs []uuid.UUID `json:"teacher_ids" validate:"required,dive,uuid"`
}
