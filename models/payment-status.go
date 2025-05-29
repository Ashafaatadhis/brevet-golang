package models

import (
	"database/sql/driver"
	"errors"
)

// PaymentStatus tipe enum
type PaymentStatus string

const (
	// Pending status
	Pending PaymentStatus = "pending"
	// Paid status
	Paid PaymentStatus = "paid"
	// Failed status
	Failed PaymentStatus = "failed"
)

// Scan implements the Scanner interface
func (ps *PaymentStatus) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan PaymentStatus: type assertion to []byte failed")
	}
	*ps = PaymentStatus(string(b))
	return nil
}

// Value implements the Valuer interface
func (ps PaymentStatus) Value() (driver.Value, error) {
	return string(ps), nil
}
