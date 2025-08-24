package models

import (
	"time"

	"github.com/google/uuid"
)

// Quiz represents a question in a quizzes
type Quiz struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MeetingID      uuid.UUID `gorm:"type:uuid;not null"`
	Title          string    `gorm:"type:text;not null"`
	Description    *string   `gorm:"type:text"`
	Type           QuizType  `gorm:"type:quiz_type;not null"`
	IsOpen         bool      `gorm:"not null;default:false"`
	StartTime      time.Time
	EndTime        time.Time
	DurationMinute int `gorm:"not null"`

	MaxAttempts int `gorm:"not null;default:1"` // 1 = hanya sekali, >1 = multi-attempt

	CreatedAt time.Time
	UpdatedAt time.Time

	Meeting   Meeting        `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	Questions []QuizQuestion `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE"`
}
