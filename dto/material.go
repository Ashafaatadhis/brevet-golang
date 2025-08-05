package dto

import (
	"time"

	"github.com/google/uuid"
)

// MaterialResponse for response
type MaterialResponse struct {
	ID          uuid.UUID `json:"id"`
	MeetingID   uuid.UUID `json:"meeting_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`

	URL string `json:"url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMaterialRequest represents the request structure for creating an material
type CreateMaterialRequest struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
	URL         string  `json:"url" validate:"required"`
}

// UpdateMaterialRequest represents the request structure for updating an material
type UpdateMaterialRequest struct {
	Title       *string `json:"title" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	URL         *string `json:"url" validate:"omitempty"`
}
