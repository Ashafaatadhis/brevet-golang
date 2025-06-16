package dto

import "github.com/google/uuid"

// GroupResponse represents the response structure for a group
type GroupResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
