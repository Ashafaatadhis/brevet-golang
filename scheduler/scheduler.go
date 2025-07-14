package scheduler

import (
	"brevet-api/config"
	"brevet-api/utils"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// StartCleanupScheduler starts a background scheduler that cleans up expired user sessions every hour
func StartCleanupScheduler(db *gorm.DB) {
	hoursStr := config.GetEnv("CLEANUP_INTERVAL_HOURS", "1")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		log.Printf("Invalid CLEANUP_INTERVAL_HOURS, fallback to 1 hour")
		hours = 1
	}

	log.Printf("Starting cleanup scheduler, interval: %d hour(s)", hours)
	ticker := time.NewTicker(time.Duration(hours) * time.Hour)
	go func() {
		for range ticker.C {
			// 1. Cleanup expired sessions
			if err := utils.CleanExpiredSessions(db); err != nil {
				log.Println("Failed to clean expired sessions:", err)
			} else {
				log.Println("Expired sessions cleaned successfully")
			}

			// 2. Mark expired purchases
			if err := utils.MarkExpiredPurchases(db); err != nil {
				log.Println("Failed to mark expired purchases:", err)
			} else {
				log.Println("Expired purchases marked successfully")
			}
		}
	}()
}
