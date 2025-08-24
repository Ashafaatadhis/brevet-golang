package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizTempSubmission represents a temporary saved answer (autosave)
type QuizTempSubmission struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AttemptID        uuid.UUID `gorm:"type:uuid;not null"`
	QuestionID       uuid.UUID `gorm:"type:uuid;not null"`
	SelectedOptionID uuid.UUID `gorm:"type:uuid;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Attempt  QuizAttempt  `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
	Question QuizQuestion `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
