package dto

import (
	"brevet-api/models"
	"time"

	"github.com/google/uuid"
)

// ResendVerificationRequest is a struct that represents the request body for resending verification code
type ResendVerificationRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyRequest represents a verification request
type VerifyRequest struct {
	Token string `json:"token" validate:"required"`
	Code  string `json:"code" validate:"required"`
}

// LoginRequest is a struct that represents the request body for logging in a user
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest is a struct that represents the request body for registering a user
type RegisterRequest struct {
	// User fields
	Name   string  `json:"name" validate:"required,min=3"`
	Phone  string  `json:"phone" validate:"required,numeric"`
	Avatar *string `json:"avatar,omitempty"`

	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`

	// Profile fields (optional)
	GroupType   models.GroupType `json:"group_type" validate:"required,group_type"`
	NIM         *string          `json:"nim,omitempty"`
	NIMProof    *string          `json:"nim_proof,omitempty"`
	NIK         *string          `json:"nik,omitempty"`
	Institution string           `json:"institution" validate:"required"`
	Origin      string           `json:"origin" validate:"required"`
	BirthDate   time.Time        `json:"birth_date" validate:"required"`
	Address     string           `json:"address" validate:"required"`
}

// RegisterResponse is a struct that represents the response body for registering a user
type RegisterResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	Avatar    *string         `json:"avatar,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	RoleType  models.RoleType `json:"role_type"`

	Profile struct {
		GroupType     *models.GroupType `json:"group_type,omitempty"`
		GroupVerified bool              `json:"group_verified"`
		NIM           *string           `json:"nim,omitempty"`
		NIMProof      *string           `json:"nim_proof,omitempty"`
		NIK           *string           `json:"nik,omitempty"`
		Institution   string            `json:"institution"`
		Origin        string            `json:"origin"`
		BirthDate     time.Time         `json:"birth_date"`
		Address       string            `json:"address"`
		CreatedAt     time.Time         `json:"created_at"`
		UpdatedAt     time.Time         `json:"updated_at"`
	} `json:"profile"`
}

// UserResponse is a struct that represents the response body for getting a user
type UserResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    *string   `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RoleType models.RoleType `json:"role_type"`

	Profile struct {
		GroupType     *models.GroupType `json:"group_type,omitempty"`
		GroupVerified bool              `json:"group_verified"`
		NIM           *string           `json:"nim,omitempty"`
		NIMProof      *string           `json:"nim_proof,omitempty"`
		NIK           *string           `json:"nik,omitempty"`
		Institution   string            `json:"institution"`
		Origin        string            `json:"origin"`
		BirthDate     time.Time         `json:"birth_date"`
		Address       string            `json:"address"`
		CreatedAt     time.Time         `json:"created_at"`
		UpdatedAt     time.Time         `json:"updated_at"`
	} `json:"profile"`
}
