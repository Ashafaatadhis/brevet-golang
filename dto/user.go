package dto

import (
	"brevet-api/models"
	"time"
)

// UpdateMyProfile is dto struct for update user
type UpdateMyProfile struct {
	// User fields
	Name *string `json:"name,omitempty" validate:"omitempty,min=3"`

	Avatar *string `json:"avatar,omitempty"`

	// Profile fields
	GroupType   *models.GroupType `json:"group_type,omitempty" validate:"omitempty,group_type"` // bisa nil
	NIM         *string           `json:"nim,omitempty" validate:"omitempty"`
	NIMProof    *string           `json:"nim_proof,omitempty" validate:"omitempty"`
	NIK         *string           `json:"nik,omitempty" validate:"omitempty"`
	Institution *string           `json:"institution,omitempty" validate:"omitempty"`
	Origin      *string           `json:"origin,omitempty" validate:"omitempty"`
	BirthDate   *time.Time        `json:"birth_date,omitempty" validate:"omitempty"` // parse manual
	Address     *string           `json:"address,omitempty" validate:"omitempty"`
}

// UpdateUserWithProfileRequest is dto struct for update user
type UpdateUserWithProfileRequest struct {
	// User fields
	Name *string `json:"name,omitempty" validate:"omitempty,min=3"`

	Avatar   *string          `json:"avatar,omitempty"`
	RoleType *models.RoleType `json:"role,omitempty" validate:"omitempty,role_type"`

	// Profile fields
	GroupType     *models.GroupType `json:"group_type,omitempty" validate:"omitempty,group_type"` // bisa nil
	GroupVerified *bool             `json:"group_verified,omitempty" validate:"omitempty"`
	NIM           *string           `json:"nim,omitempty" validate:"omitempty"`
	NIMProof      *string           `json:"nim_proof,omitempty" validate:"omitempty"`
	NIK           *string           `json:"nik,omitempty" validate:"omitempty"`
	Institution   *string           `json:"institution,omitempty" validate:"omitempty"`
	Origin        *string           `json:"origin,omitempty" validate:"omitempty"`
	BirthDate     *time.Time        `json:"birth_date,omitempty" validate:"omitempty"` // parse manual
	Address       *string           `json:"address,omitempty" validate:"omitempty"`
}

// CreateUserWithProfileRequest is dto struct for create user with profile
type CreateUserWithProfileRequest struct {
	Name   string  `json:"name" validate:"required,min=3"`
	Phone  string  `json:"phone" validate:"required,numeric"`
	Avatar *string `json:"avatar,omitempty"`

	Email           string          `json:"email" validate:"required,email"`
	Password        string          `json:"password" validate:"required,min=6"`
	ConfirmPassword string          `json:"confirm_password" validate:"required,eqfield=Password"`
	RoleType        models.RoleType `json:"role_type" validate:"required,role_type"`

	// Profile fields (optional)
	GroupType   *models.GroupType `json:"group_type" validate:"omitempty,group_type"`
	NIM         *string           `json:"nim,omitempty"`
	NIMProof    *string           `json:"nim_proof,omitempty"`
	NIK         *string           `json:"nik,omitempty"`
	Institution string            `json:"institution" validate:"required"`
	Origin      string            `json:"origin" validate:"required"`
	BirthDate   time.Time         `json:"birth_date" validate:"required"`
	Address     string            `json:"address" validate:"required"`
}
