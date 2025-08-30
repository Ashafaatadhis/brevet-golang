package services

import (
	"brevet-api/config"
	"brevet-api/repository"
	"context"

	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IVerificationService interface
type IVerificationService interface {
	GenerateVerificationCode(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (string, error)
	VerifyCode(ctx context.Context, userID uuid.UUID, code string) bool
	GetCooldownRemaining(ctx context.Context, userID uuid.UUID) (time.Duration, error)
}

// VerificationService is a service for handling user verification codes
type VerificationService struct {
	repo repository.IVerificationRepository
}

// NewVerificationService creates a new instance of VerificationService
func NewVerificationService(repo repository.IVerificationRepository) IVerificationService {
	return &VerificationService{repo: repo}
}

// GenerateVerificationCode generates a new verification code for a user
func (s *VerificationService) GenerateVerificationCode(ctx context.Context, tx *gorm.DB, userID uuid.UUID) (string, error) {
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

	if tx == nil {
		if err := s.repo.UpdateVerificationCode(ctx, userID, codeStr, expiry); err != nil {
			return "", err
		}
	} else {
		if err := s.repo.WithTx(tx).UpdateVerificationCode(ctx, userID, codeStr, expiry); err != nil {
			return "", err
		}
	}

	return codeStr, nil
}

// VerifyCode checks if the provided verification code is valid for the user
func (s *VerificationService) VerifyCode(ctx context.Context, userID uuid.UUID, code string) bool {
	user, err := s.repo.FindUserByCode(ctx, userID, code)
	if err != nil {
		return false
	}

	if err := s.repo.MarkUserVerified(ctx, user.ID); err != nil {
		return false
	}

	return true
}

// GetCooldownRemaining returns the remaining cooldown time for sending a new verification code
func (s *VerificationService) GetCooldownRemaining(ctx context.Context, userID uuid.UUID) (time.Duration, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
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
