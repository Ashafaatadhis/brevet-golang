package services

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService is a struct that represents a user service
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetAllFilteredUsers is a method that returns all users
func (s *UserService) GetAllFilteredUsers(opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields, err := utils.GetValidColumns(s.db, &models.User{}, &models.Profile{})

	if err != nil {
		return nil, 0, err
	}

	sort := opts.Sort
	if !validSortFields[sort] {

		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := s.db.Model(&models.User{})

	// Define custom join conditions for relations
	joinConditions := map[string]string{
		"profile": "LEFT JOIN profiles AS profiles ON profiles.user_id = users.id",
		"role":    "LEFT JOIN roles AS roles ON roles.id = users.role_id",
	}

	joinedRelations := map[string]bool{}

	for key, val := range opts.Filters {
		if strings.Contains(key, ".") {
			// Validate relation.column filter
			if !validSortFields[key] {
				continue // skip invalid filter
			}
			parts := strings.SplitN(key, ".", 2)
			relation, column := parts[0], parts[1]
			alias := relation + "s"
			if !joinedRelations[relation] {
				if cond, ok := joinConditions[relation]; ok {
					db = db.Joins(cond)
				} else {
					db = db.Joins(fmt.Sprintf("LEFT JOIN %ss AS %s ON %s.id = users.%s_id", relation, alias, alias, relation))
				}
				joinedRelations[relation] = true
			}
			db = db.Where(fmt.Sprintf("%s.%s = ?", alias, column), val)
		} else {
			// Validate direct column filter
			if validSortFields[key] {
				db = db.Where(fmt.Sprintf("%s = ?", key), val)
			}
		}
	}

	if opts.Search != "" {
		db = db.Where("name ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var users []models.User
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("Role").
		Preload("Profile").
		Find(&users).Error

	return users, total, err
}

// GetUserByID is a method that returns a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Role").Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
