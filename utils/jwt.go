package utils

import (
	"brevet-api/config"
	"brevet-api/models"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Username string    `json:"username,omitempty"`
	Name     string    `json:"name,omitempty"`
	jwt.RegisteredClaims
}

// VerificationClaims adalah struktur claims JWT untuk verifikasi kode
type VerificationClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateVerificationToken membuat JWT token untuk verifikasi user dengan kode 6 digit
func GenerateVerificationToken(userID uuid.UUID, email string) (string, error) {
	expiryStr := config.GetEnv("VERIFICATION_TOKEN_EXPIRY_MINUTES", "15")
	jwtSecret := config.GetEnv("VERIFICATION_TOKEN_SECRET", "default-key")

	expiryMinutes, err := strconv.Atoi(expiryStr)
	if err != nil {
		expiryMinutes = 15
	}

	expirationTime := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)

	claims := &VerificationClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// GenerateJWT generates a JWT token for the given user
func GenerateJWT(user models.User, jwtSecret string, expiryHours int) (string, error) {

	expirationTime := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.RoleType), // Make sure Role.Name is always loaded
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ExtractUserIDFromToken extracts the user ID from a JWT token
func ExtractUserIDFromToken(tokenString string, jwtSecret string) (*VerificationClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &VerificationClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})

	claims, ok := token.Claims.(*VerificationClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Kalau error karena expired, tetap kembalikan claims
	if errors.Is(err, jwt.ErrTokenExpired) {
		return claims, nil
	}

	// Kalau error lainnya, tolak
	if err != nil {
		return nil, err
	}

	// Kalau valid dan tidak error
	if token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}

// ExtractToken extracts claims from a JWT token and returns the claims as a map
func ExtractToken(tokenString string, jwtSecret string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}

// ExtractClaimsFromToken extracts the Claims struct from a JWT token
func ExtractClaimsFromToken(tokenString string, jwtSecret string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// If token is expired, return error
	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, err
	}

	// Other errors
	if err != nil {
		return nil, err
	}

	// Valid token
	if token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid or expired token")
}
