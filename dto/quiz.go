package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// ImportQuizzesRequest request
type ImportQuizzesRequest struct {
	Title          string          `json:"title" validate:"required"`
	Description    *string         `json:"description" validate:"omitempty"`
	Type           models.QuizType `json:"quiz_type" validate:"required,quiz_type"`
	DurationMinute int             `json:"duration_minute" validate:"required,min=1"`
	MaxAttempts    int             `json:"max_attempts" validate:"required"`
	IsOpen         bool            `json:"is_open" validate:"required"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
}

// SaveTempSubmissionRequest request
type SaveTempSubmissionRequest struct {
	QuestionID       uuid.UUID `json:"question_id" validate:"required"`
	SelectedOptionID uuid.UUID `json:"selected_option_id" validate:"required"`
}

// UpdateQuizRequest request
type UpdateQuizRequest struct {
	Title          *string          `json:"title,omitempty"`
	Description    *string          `json:"description,omitempty"`
	Type           *models.QuizType `json:"type,omitempty"`
	IsOpen         *bool            `json:"is_open,omitempty"`
	StartTime      *time.Time       `json:"start_time,omitempty"`
	EndTime        *time.Time       `json:"end_time,omitempty"`
	DurationMinute *int             `json:"duration_minute,omitempty"`
	MaxAttempts    *int             `json:"max_attempts,omitempty"`
}

// QuizResponse response
type QuizResponse struct {
	ID             uuid.UUID       `json:"id"`
	MeetingID      uuid.UUID       `json:"meeting_id"`
	Title          string          `json:"title"`
	Description    *string         `json:"description"`
	Type           models.QuizType `json:"type"`
	IsOpen         bool            `json:"is_open"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	DurationMinute int             `json:"duration_minute"`

	MaxAttempts int `json:"max_attempts"`

	Questions []QuestionResponse `json:"questions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizResultResponse response
type QuizResultResponse struct {
	ID        uuid.UUID `json:"id"`
	AttemptID uuid.UUID `json:"attempt_id"`

	TotalQuestions int     `json:"total_questions"`
	CorrectAnswers int     `json:"correct_answers"`
	WrongAnswers   int     `json:"wrong_answers"`
	ScorePercent   float64 `json:"score_percent"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Attempt QuizAttemptWithQuizMetadataResponse `json:"attempt"`
}

// QuestionResponse response
type QuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	QuizID   uuid.UUID `json:"quiz_id"`
	Question string    `json:"question"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Options  []QuizOptionResponse         `json:"options,omitempty"`
	TempSubs []QuizTempSubmissionResponse `json:"temp_subs,omitempty"`
	// Subs     []QuizSubmission     `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}

// QuizOptionResponse response
type QuizOptionResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	OptionText string    `json:"option_text"`
	IsCorrect  bool      `json:"is_correct"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizTempSubmissionResponse response
type QuizTempSubmissionResponse struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	QuestionID       uuid.UUID `json:"question_id"`
	SelectedOptionID uuid.UUID `json:"selected_option_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizAttemptResponse response
type QuizAttemptResponse struct {
	ID        uuid.UUID  `json:"id"`
	QuizID    uuid.UUID  `json:"quiz_id"`
	UserID    uuid.UUID  `json:"user_id"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizAttemptWithQuizMetadataResponse response
type QuizAttemptWithQuizMetadataResponse struct {
	ID        uuid.UUID  `json:"id"`
	QuizID    uuid.UUID  `json:"quiz_id"`
	UserID    uuid.UUID  `json:"user_id"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`

	Quiz QuizForUserResponse `json:"quiz"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizOptionForUserResponse hanya untuk user, tanpa IsCorrect
type QuizOptionForUserResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"question_id"`
	OptionText string    `json:"option_text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuestionForUserResponse hanya untuk user, tanpa jawaban
type QuestionForUserResponse struct {
	ID       uuid.UUID `json:"id"`
	QuizID   uuid.UUID `json:"quiz_id"`
	Question string    `json:"question"`

	Options  []QuizOptionForUserResponse  `json:"options,omitempty"`
	TempSubs []QuizTempSubmissionResponse `json:"temp_subs,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizForUserResponse hanya untuk user, tanpa jawaban
type QuizForUserResponse struct {
	ID             uuid.UUID       `json:"id"`
	MeetingID      uuid.UUID       `json:"meeting_id"`
	Title          string          `json:"title"`
	Description    *string         `json:"description"`
	Type           models.QuizType `json:"type"`
	IsOpen         bool            `json:"is_open"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	DurationMinute int             `json:"duration_minute"`

	MaxAttempts int `json:"max_attempts"`

	Questions []QuestionForUserResponse `json:"questions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuizAttemptFullResponse response
type QuizAttemptFullResponse struct {
	Attempt         *QuizAttemptResponse          `json:"attempt"`
	Quiz            *QuizForUserResponse          `json:"quiz"`
	TempSubmissions []*QuizTempSubmissionResponse `json:"temp_submissions,omitempty"`
}

// QuizAttemptFull response
type QuizAttemptFull struct {
	Attempt         *models.QuizAttempt         `json:"attempt"`
	Quiz            *models.Quiz                `json:"quiz"`
	TempSubmissions []models.QuizTempSubmission `json:"temp_submissions,omitempty"`
}
