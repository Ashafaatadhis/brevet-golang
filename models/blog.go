package models

import "github.com/google/uuid"

// Blog represents the blog data model
type Blog struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Slug        string    `gorm:"type:varchar(255);not null;unique"`
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text;not null"`
	Content     string    `gorm:"type:text;not null"`
	Image       string    `gorm:"type:varchar(255)"`
}
