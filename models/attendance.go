package models

import (
	"time"

	"github.com/google/uuid"
)

// Attendance is a struct that represents a attendance
type Attendance struct {
	ID        uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MeetingID uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_user"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_meeting_user"`
	Status    AttendanceStatus `gorm:"type:attendance_status;not null"`
	Note      *string          `gorm:"type:text"`

	UpdatedBy uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Meeting       Meeting `gorm:"foreignKey:MeetingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User          User    `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UpdatedByUser User    `gorm:"foreignKey:UpdatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
