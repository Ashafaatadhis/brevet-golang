package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizSubmission represents a final submitted answer
type QuizSubmission struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AttemptID        uuid.UUID `gorm:"type:uuid;not null"`
	QuestionID       uuid.UUID `gorm:"type:uuid;not null"`
	SelectedOptionID uuid.UUID `gorm:"type:uuid;not null"`
	Score            int       `gorm:"not null"` // 1 = correct, 0 = wrong

	CreatedAt time.Time
	UpdatedAt time.Time

	Attempt  QuizAttempt  `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
	Question QuizQuestion `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
