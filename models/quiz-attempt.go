package models

import (
	"time"

	"github.com/google/uuid"
)

// QuizAttempt mencatat saat user mulai quiz dan durasi attempt
type QuizAttempt struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	QuizID    uuid.UUID  `gorm:"type:uuid;not null"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	StartedAt time.Time  `gorm:"not null"`     // Waktu user mulai quiz
	EndedAt   *time.Time `gorm:"default:null"` // Waktu selesai, diisi saat submit

	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Quiz Quiz `gorm:"foreignKey:QuizID;constraint:OnDelete:CASCADE"`

	Submissions []QuizSubmission `gorm:"foreignKey:AttemptID"`
}
