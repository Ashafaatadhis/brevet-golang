package dto

import "github.com/google/uuid"

// BlogResponse represents the response structure for a blog
type BlogResponse struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Image       string    `json:"image"`
}

// CreateBlogRequest represents the request structure for creating a blog
type CreateBlogRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Content     string `json:"content" validate:"required"`
	Image       string `json:"image" validate:"required"`
}

// UpdateBlogRequest represents the request structure for updating a blog
type UpdateBlogRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty"`
	Description *string `json:"description,omitempty" validate:"omitempty"`
	Content     *string `json:"content,omitempty" validate:"omitempty"`
	Image       *string `json:"image,omitempty" validate:"omitempty"`
}
