package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateTestimonialRequest request
type CreateTestimonialRequest struct {
	Rating      int     `json:"rating" validate:"required,min=1,max=5"`
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description" validate:"omitempty"`
}

// UpdateTestimonialRequest request
type UpdateTestimonialRequest struct {
	Rating      *int    `json:"rating" validate:"omitempty,min=1,max=5"`
	Title       *string `json:"title" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
}

// TestimonialResponse response
type TestimonialResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	BatchID     uuid.UUID `json:"batch_id"`
	Rating      int       `json:"rating"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User  UserResponse  `json:"user"`
	Batch BatchResponse `json:"batch"`
}
