package utils

import (
	"time"

	"gorm.io/gorm"
)

// MarkExpiredPurchases updates all pending purchases that have passed their expired_at time
func MarkExpiredPurchases(db *gorm.DB) error {
	return db.Exec(`
		UPDATE purchases
		SET payment_status = 'expired', updated_at = NOW()
		WHERE payment_status = 'pending'
		AND expired_at IS NOT NULL
		AND expired_at <= ?
	`, time.Now()).Error
}
