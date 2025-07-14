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

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMeetingRequest is request income
type CreateMeetingRequest struct {
	Title       string             `json:"title" validate:"required"`
	Description string             `json:"description"`
	Type        models.MeetingType `json:"type" validate:"required,meeting_type"`
}

// AssignTeachersRequest AssignTeachersRequest is request income
type AssignTeachersRequest struct {
	TeacherIDs []uuid.UUID `json:"teacher_ids" validate:"required,dive,uuid"`
}
