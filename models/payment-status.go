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
	// WaitingConfirmation status
	WaitingConfirmation PaymentStatus = "waiting_confirmation"
	// Paid status
	Paid PaymentStatus = "paid"
	// Rejected status
	Rejected PaymentStatus = "rejected"
	// Expired status
	Expired PaymentStatus = "expired"
	// Cancelled status
	Cancelled PaymentStatus = "cancelled"
)

// Scan implements the Scanner interface
func (ps *PaymentStatus) Scan(value any) error {

	switch v := value.(type) {
	case []byte:
		*ps = PaymentStatus(string(v))
		return nil
	case string:
		*ps = PaymentStatus(v)
		return nil
	}
	return errors.New("failed to scan PaymentStatus: invalid type")

}

// Value implements the Valuer interface
func (ps PaymentStatus) Value() (driver.Value, error) {
	return string(ps), nil
}
