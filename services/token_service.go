package services

import (
	"brevet-api/models"
	"brevet-api/utils"

	"github.com/google/uuid"
)

// ITokenService servjce
type ITokenService interface {
	ExtractUserIDFromToken(token, secret string) (*utils.VerificationClaims, error)
	GenerateVerificationToken(userID uuid.UUID, email string) (string, error)
	GenerateJWT(user models.User, jwtSecret string, expiryHours int) (string, error)
}

// TokenService struct
type TokenService struct {
}

// NewTokenService init
func NewTokenService() ITokenService {
	return &TokenService{}
}

// ExtractUserIDFromToken for extract from called utils
func (t *TokenService) ExtractUserIDFromToken(token, secret string) (*utils.VerificationClaims, error) {
	return utils.ExtractUserIDFromToken(token, secret)
}

// GenerateVerificationToken for generate from called utils
func (t *TokenService) GenerateVerificationToken(userID uuid.UUID, email string) (string, error) {
	return utils.GenerateVerificationToken(userID, email)
}

// GenerateJWT for generate from called utils
func (t *TokenService) GenerateJWT(user models.User, jwtSecret string, expiryHours int) (string, error) {
	return utils.GenerateJWT(user, jwtSecret, expiryHours)
}
