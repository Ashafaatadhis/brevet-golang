package models

import (
	"time"

	"github.com/google/uuid"
)

// Testimonial for models
type Testimonial struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_user_batch,unique"`
	BatchID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_batch,unique"`
	Rating      int       `gorm:"not null"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Batch Batch `gorm:"foreignKey:BatchID;constraint:OnDelete:CASCADE"`
}
