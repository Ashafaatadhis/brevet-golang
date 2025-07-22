package utils

import (
	"gorm.io/gorm"
)

// WithTransaction is a utility function to execute a function within a transaction.
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // lempar ulang panic agar tetap masuk Fiber ErrorHandler
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	err = fn(tx)
	return
}
