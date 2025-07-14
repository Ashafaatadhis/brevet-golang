package models

import (
	"time"

	"github.com/google/uuid"
)

// MeetingTeacher is pivot table
type MeetingTeacher struct {
	MeetingID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time

	// Optional relasi balik (biar bisa preload)
	Meeting Meeting `gorm:"foreignKey:MeetingID;constraint:OnDelete:CASCADE"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
