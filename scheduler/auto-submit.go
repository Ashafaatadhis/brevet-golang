package scheduler

import (
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/services"
	"brevet-api/utils"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InitQuizScheduler inisialisasi dependency quiz + jalanin scheduler
func InitQuizScheduler(db *gorm.DB) {
	emailService, err := services.NewEmailServiceFromEnv()
	if err != nil {
		panic(err)
	}

	userRepository := repository.NewUserRepository(db)
	fileService := services.NewFileService()
	batchRepository := repository.NewBatchRepository(db)
	meetingRepo := repository.NewMeetingRepository(db)

	purchaseRepo := repository.NewPurchaseRepository(db)
	purchaseService := services.NewPurchaseService(
		purchaseRepo, userRepository, batchRepository, emailService, db,
	)

	quizRepository := repository.NewQuizRepository(db)
	quizService := services.NewQuizService(
		quizRepository, batchRepository, meetingRepo, purchaseService, fileService, db,
	)

	go startAutoSubmitScheduler(db, quizService)
}

func startAutoSubmitScheduler(db *gorm.DB, quizService services.IQuizService) {
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

				// Hitung endTime attempt
				endTime := attempt.StartedAt.Add(time.Duration(quiz.DurationMinute) * time.Minute)
				if !quiz.EndTime.IsZero() && quiz.EndTime.Before(endTime) {
					endTime = quiz.EndTime
				}

				if now.After(endTime) {
					// Auto-submit pakai service biar logikanya sama dengan manual submit
					user := &utils.Claims{UserID: attempt.UserID} // bikin claim dummy
					if err := quizService.SubmitQuiz(context.Background(), user, attempt.ID); err != nil {
						fmt.Println("Failed to auto-submit attempt:", attempt.ID, err)
					} else {
						fmt.Println("Auto-submitted attempt:", attempt.ID)
					}
				}
			}
		}
	}()
}
