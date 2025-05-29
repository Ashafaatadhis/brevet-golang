package services

import (
	"brevet-api/config"
	"brevet-api/models"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VerificationService handles verification-related operations
type VerificationService struct {
	db *gorm.DB
}

// NewVerificationService creates a new VerificationService
func NewVerificationService(db *gorm.DB) *VerificationService {
	return &VerificationService{db: db}
}

// GenerateVerificationCode generates a 6-digit verification code for a user
func (s *VerificationService) GenerateVerificationCode(userID uuid.UUID) (string, error) {
	// Generate 6-digit random code
	code := rand.Intn(900000) + 100000
	codeStr := fmt.Sprintf("%06d", code)

	// Ambil expiry dari env, fallback ke 15 menit
	expiryStr := config.GetEnv("VERIFICATION_EXPIRY_MINUTES", "15")
	expiryMinutes, err := strconv.Atoi(expiryStr)
	if err != nil || expiryMinutes <= 0 {
		expiryMinutes = 15
	}
	expiry := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	// Update database: code, expiry, dan waktu terakhir dikirim
	if err := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"verify_code":  codeStr,
			"code_expiry":  expiry,
			"last_sent_at": time.Now(),
		}).Error; err != nil {
		return "", err
	}

	return codeStr, nil
}

// VerifyCode verifies a verification code for a user
func (s *VerificationService) VerifyCode(userID uuid.UUID, code string) bool {
	var user models.User
	if err := s.db.Where("id = ? AND verify_code = ? AND code_expiry > ?",
		userID, code, time.Now()).First(&user).Error; err != nil {
		return false
	}

	// Mark as verified and clear verification fields
	s.db.Model(&user).Updates(map[string]any{
		"is_verified": true,
		"verify_code": nil,
		"code_expiry": nil,
	})

	return true
}

// GetCooldownRemaining returns the remaining cooldown time for a user
func (s *VerificationService) GetCooldownRemaining(userID uuid.UUID) (time.Duration, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}

	if !user.LastSentAt.Valid {
		return 0, nil
	}

	nextAllowed := user.LastSentAt.Time.Add(2 * time.Minute)
	remaining := time.Until(nextAllowed)
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}
