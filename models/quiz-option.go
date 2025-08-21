package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizOption represents an option for a quiz
type QuizOption struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	QuestionID uuid.UUID `gorm:"type:uuid;not null;index"`
	OptionText string    `gorm:"type:text;not null"`
	IsCorrect  bool      `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Question QuizQuestion `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
