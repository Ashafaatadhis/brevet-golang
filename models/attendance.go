package models

import (
	"time"

	"github.com/google/uuid"
)

// Attendance is a struct that represents a attendance
type Attendance struct {
	ID             uuid.UUID        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MeetingID      uuid.UUID        `gorm:"type:uuid;not null"`
	UserID         uuid.UUID        `gorm:"type:uuid;not null"`
	Status         AttendanceStatus `gorm:"type:attendance_status;not null"`
	Note           string           `gorm:"type:text"`
	AttendanceTime time.Time        `gorm:"type:timestamp"`
	UpdatedBy      uuid.UUID        `gorm:"type:uuid"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Meeting       Meeting `gorm:"foreignKey:MeetingID"`
	User          User    `gorm:"foreignKey:UserID"`
	UpdatedByUser User    `gorm:"foreignKey:UpdatedBy"`
}
