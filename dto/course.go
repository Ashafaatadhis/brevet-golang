package dto

import (
	"time"

	"github.com/google/uuid"
)

// CourseResponse represents the response structure for a course
type CourseResponse struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	ShortDescription string    `json:"short_description"`
	Description      string    `json:"description"`
	LearningOutcomes string    `json:"learning_outcomes"`
	Achievements     string    `json:"achievements"`

	// CourseThumbnail string `gorm:"type:varchar"`
	CourseImages []CourseImageResponse `json:"course_images"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CourseImageResponse represents the response structure for a course image
type CourseImageResponse struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCourseRequest represents the request structure for creating a course
type CreateCourseRequest struct {
	Title            string               `json:"title" validate:"required"`
	ShortDescription string               `json:"short_description" validate:"required"`
	Description      string               `json:"description" validate:"required"`
	LearningOutcomes string               `json:"learning_outcomes" validate:"required"`
	Achievements     string               `json:"achievements" validate:"required"`
	CourseImages     []CourseImageRequest `json:"course_images" validate:"required,dive"`
}

// UpdateCourseRequest represents the request structure for updating a course
type UpdateCourseRequest struct {
	Title            *string               `json:"title,omitempty" validate:"omitempty"`
	ShortDescription *string               `json:"short_description,omitempty" validate:"omitempty"`
	Description      *string               `json:"description,omitempty" validate:"omitempty"`
	LearningOutcomes *string               `json:"learning_outcomes,omitempty" validate:"omitempty"`
	Achievements     *string               `json:"achievements,omitempty" validate:"omitempty"`
	CourseImages     *[]CourseImageRequest `json:"course_images,omitempty" validate:"omitempty,dive"`
}

// CourseImageRequest represents the request structure for a course image
type CourseImageRequest struct {
	ImageURL string `json:"image_url" validate:"required"`
}
