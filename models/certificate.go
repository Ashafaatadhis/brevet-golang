package models

import (
	"time"

	"github.com/google/uuid"
)

// Certificate for table certificates
type Certificate struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID   uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Number    string    `gorm:"type:varchar(50);uniqueIndex;not null"` // nomor sertifikat resmi
	URL       string    `gorm:"type:text;not null"`
	QRCode    string    `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
