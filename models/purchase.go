package models

import (
	"time"

	"github.com/google/uuid"
)

// Purchase is model for table purchases
type Purchase struct {
	ID            uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	PaymentStatus PaymentStatus `gorm:"type:payment_status;not null"`

	UserID *uuid.UUID `gorm:"type:uuid"`
	User   *User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:SET NULL"`

	BatchID *uuid.UUID `gorm:"type:uuid"`
	Batch   *Batch     `gorm:"foreignKey:BatchID;references:ID;constraint:OnDelete:SET NULL"`

	PriceID uuid.UUID `gorm:"type:uuid;not null"`
	Price   Price     `gorm:"foreignKey:PriceID;references:ID"`

	PaymentProof *string    `gorm:"type:varchar(255)"`
	ExpiredAt    *time.Time `gorm:"type:timestamp"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
