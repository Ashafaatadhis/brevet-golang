package dto

import (
	"time"

	"github.com/google/uuid"
)

// CertificateResponse response
type CertificateResponse struct {
	ID        uuid.UUID `json:"id"`
	BatchID   uuid.UUID `json:"batch_id"`
	UserID    uuid.UUID `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	URL       string    `json:"url"`
	QRCode    string    `json:"qr_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
