package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// AttendanceResponse is response struct
type AttendanceResponse struct {
	ID             uuid.UUID               `json:"id"`
	MeetingID      uuid.UUID               `json:"meeting_id"`
	UserID         uuid.UUID               `json:"user_id"`
	Status         models.AttendanceStatus `json:"attendance_status"`
	Note           *string                 `json:"note"`
	AttendanceTime time.Time               `json:"attendance_time"`
	UpdatedBy      uuid.UUID               `json:"updated_by"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// BulkAttendanceItem represent request for
type BulkAttendanceItem struct {
	UserID uuid.UUID               `json:"user_id" validate:"required"`
	Status models.AttendanceStatus `json:"status" validate:"required,attendance_status"`
	Note   *string                 `json:"note"`
}

// BulkAttendanceRequest represent request for
type BulkAttendanceRequest struct {
	Attendances []BulkAttendanceItem `json:"attendances" validate:"required,dive"`
}
