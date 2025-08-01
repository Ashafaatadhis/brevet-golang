package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// BatchResponse represents the response structure for a batch
type BatchResponse struct {
	ID             uuid.UUID `json:"id"`
	CourseID       uuid.UUID `json:"course_id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    *string   `json:"description"`
	BatchThumbnail *string   `json:"batch_thumbnail"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	StartTime      string    `json:"start_time"` // HH:mm
	EndTime        string    `json:"end_time"`
	Room           string    `json:"room"`
	Quota          int       `json:"quota"`
	Days           []*struct {
		ID      uuid.UUID      `json:"id"`
		BatchID uuid.UUID      `json:"batch_id"`
		Day     models.DayType `json:"day"`
	} `json:"days"`

	BatchGroups []struct {
		ID      uuid.UUID `json:"id"`
		BatchID uuid.UUID `json:"batch_id"`

		GroupType models.GroupType `json:"group_type"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"batch_groups"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CourseType models.CourseType `json:"course_type"`
}

// CreateBatchRequest represents the request structure for creating a batch
type CreateBatchRequest struct {
	Title          string            `json:"title" validate:"required"`
	Description    *string           `json:"description" validate:"omitempty"`
	BatchThumbnail *string           `json:"batch_thumbnail,omitempty"`
	StartAt        time.Time         `json:"start_at" validate:"required"`
	EndAt          time.Time         `json:"end_at" validate:"required"`
	StartTime      string            `json:"start_time" validate:"required,datetime=15:04"`
	EndTime        string            `json:"end_time" validate:"required,datetime=15:04"`
	Days           []models.DayType  `json:"days" validate:"required,min=1,dive,required,day_type"`
	Room           string            `json:"room" validate:"required"`
	Quota          int               `json:"quota" validate:"required,min=1"`
	CourseType     models.CourseType `json:"course_type" validate:"required,course_type"`

	GroupTypes []models.GroupType `json:"group_types" validate:"required,min=1,dive,required,group_type"`
}

// UpdateBatchRequest represents the request structure for updating a batch
type UpdateBatchRequest struct {
	Title          *string            `json:"title,omitempty" validate:"omitempty"`
	Description    *string            `json:"description,omitempty" validate:"omitempty"`
	BatchThumbnail *string            `json:"batch_thumbnail,omitempty" validate:"omitempty"`
	StartAt        *time.Time         `json:"start_at,omitempty" validate:"omitempty"`
	EndAt          *time.Time         `json:"end_at,omitempty" validate:"omitempty"`
	StartTime      *string            `json:"start_time" validate:"omitempty,datetime=15:04"` // HH:mm
	EndTime        *string            `json:"end_time" validate:"omitempty,datetime=15:04"`   // HH:mm
	Days           *[]models.DayType  `json:"days,omitempty" validate:"omitempty,min=1,dive,required,day_type"`
	Room           *string            `json:"room,omitempty" validate:"omitempty"`
	Quota          *int               `json:"quota,omitempty" validate:"omitempty,min=1"`
	CourseType     *models.CourseType `json:"course_type,omitempty" validate:"omitempty,course_type"`

	GroupTypes *[]models.GroupType `json:"group_types,omitempty" validate:"omitempty,min=1,dive,required,group_type"`
}

// BATCH TEACHER

// CreateBatchTeacherRequest represents the request structure for creating a batch teacher
type CreateBatchTeacherRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}
