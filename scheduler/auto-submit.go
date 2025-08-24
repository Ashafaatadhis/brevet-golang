package scheduler

import (
	"brevet-api/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StartAutoSubmitScheduler starts the auto-submit scheduler for a quiz
func StartAutoSubmitScheduler(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			var attempts []models.QuizAttempt
			if err := db.Where("ended_at IS NULL").Find(&attempts).Error; err != nil {
				fmt.Println("Failed to fetch attempts:", err)
				continue
			}

			for _, attempt := range attempts {
				var quiz models.Quiz
				if err := db.First(&quiz, "id = ?", attempt.QuizID).Error; err != nil {
					continue
				}

				// hitung waktu berakhir attempt
				endTime := attempt.StartedAt.Add(time.Duration(quiz.DurationMinute) * time.Minute)
				if !quiz.EndTime.IsZero() && quiz.EndTime.Before(endTime) {
					endTime = quiz.EndTime
				}

				if now.After(endTime) {
					attempt.EndedAt = &now
					if err := db.Save(&attempt).Error; err != nil {
						fmt.Println("Failed to auto-submit attempt:", attempt.ID, err)
					} else {
						fmt.Println("Auto-submitted attempt:", attempt.ID)
					}
				}
			}
		}
	}()
}
