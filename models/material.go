package models

import (
	"time"

	"github.com/google/uuid"
)

// Material is struct model
type Material struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	MeetingID   uuid.UUID `gorm:"type:uuid;not null"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`

	URL string `gorm:"type:text;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Meeting Meeting `gorm:"foreignKey:MeetingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
