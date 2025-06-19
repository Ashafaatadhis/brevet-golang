package dto

import (
	"brevet-api/models"
	"time"
)

// UpdateUserWithProfileRequest is dto struct for update user
type UpdateUserWithProfileRequest struct {
	// User fields
	Name *string `json:"name,omitempty" validate:"omitempty"`

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
