package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizResult menyimpan hasil akhir dari satu attempt
type QuizResult struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AttemptID uuid.UUID `gorm:"type:uuid;not null;index"` // relasi ke QuizAttempt

	TotalQuestions int     `gorm:"not null"`
	CorrectAnswers int     `gorm:"not null"`
	WrongAnswers   int     `gorm:"not null"`
	ScorePercent   float64 `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Attempt QuizAttempt `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
}
