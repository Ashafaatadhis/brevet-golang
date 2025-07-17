package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// CreatePurchase for body createpurchase
type CreatePurchase struct {
	BatchID uuid.UUID `json:"batch_id"`
}

// PurchaseResponse for struct response
type PurchaseResponse struct {
	ID            uuid.UUID            `json:"id"`
	PaymentStatus models.PaymentStatus `json:"payment_status"`
	UserID        uuid.UUID            `json:"user_id"`
	BatchID       uuid.UUID            `json:"batch_id"`
	PriceID       uuid.UUID            `json:"price_id"`
	ExpiredAt     *time.Time           `json:"expired_at"`

	User *UserResponse `json:"user,omitempty"`

	Batch *BatchResponse `json:"batch,omitempty"`

	Price *struct {
		ID        uuid.UUID        `json:"id"`
		GroupType models.GroupType `json:"group_type"`

		Price float64 `json:"price"`

		CreatedAt time.Time `json:"updated_at"`
		UpdatedAt time.Time `json:"created_at"`
	} `json:"price,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PayPurchaseRequest struct for pay purchase
type PayPurchaseRequest struct {
	PaymentProofURL string `json:"payment_proof_url" validate:"required"`
}

// UpdateStatusPayment struct for update status payment
type UpdateStatusPayment struct {
	PaymentStatus models.PaymentStatus `json:"payment_status" validate:"required,payment_status_type"`
}
