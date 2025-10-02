package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// ScoreResponse response
type ScoreResponse struct {
	Assignments []AssignmentScore `json:"assignments"`
	Quizzes     []QuizScore       `json:"quizzes"`
}

// AssignmentScore response
type AssignmentScore struct {
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
	Score       float64               `json:"score"`
}

// QuizScore response
type QuizScore struct {
	ID uuid.UUID `json:"id"`

	MeetingID      uuid.UUID       `json:"meeting_id"`
	Title          string          `json:"title"`
	Description    *string         `json:"description"`
	Type           models.QuizType `json:"type"`
	IsOpen         bool            `json:"is_open"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	DurationMinute int             `json:"duration_minute"`

	MaxAttempts int `json:"max_attempts"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Score     float64   `json:"score"`
}
