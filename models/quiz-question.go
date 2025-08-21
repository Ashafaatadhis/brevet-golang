package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizQuestion adalah soal individual di dalam satu Quiz
type QuizQuestion struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	QuizID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Question string    `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Quiz     Quiz                 `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE"`
	Options  []QuizOption         `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
	TempSubs []QuizTempSubmission `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
	Subs     []QuizSubmission     `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE"`
}
