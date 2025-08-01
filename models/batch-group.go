package models

import (
	"time"

	"github.com/google/uuid"
)

// BatchGroup is a struct that represents a batch group
type BatchGroup struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BatchID uuid.UUID `gorm:"type:uuid;not null"`
	Batch   Batch     `gorm:"foreignKey:BatchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	GroupType GroupType `gorm:"type:group_type;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
