package models

import (
	"time"

	"github.com/google/uuid"
)

// Purchase is a struct that represents a purchase
type Purchase struct {
	ID            uuid.UUID     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	PaymentStatus PaymentStatus `gorm:"type:payment_status;not null"`

	UserID uuid.UUID `gorm:"type:uuid;not null"`
	User   User      `gorm:"foreignKey:UserID;references:ID"`

	BatchID uuid.UUID `gorm:"type:uuid;not null"`
	Batch   Batch     `gorm:"foreignKey:BatchID;references:ID"`

	PriceID uuid.UUID `gorm:"type:uuid;not null"`
	Price   Price     `gorm:"foreignKey:PriceID;references:ID"`

	ConfirmationURL string `gorm:"type:varchar(255)"`
	PaymentProof    string `gorm:"type:varchar(255)"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
