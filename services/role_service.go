package services

import (
	"brevet-api/models"

	"gorm.io/gorm"
)

// RoleService is a service for role-related operations
type RoleService struct {
	db *gorm.DB
}

// NewRoleService creates a new RoleService instance
func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// GetRoleByName retrieves a role by its name
func (s *RoleService) GetRoleByName(db *gorm.DB, name string) (*models.Role, error) {
	var role models.Role
	if err := db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
