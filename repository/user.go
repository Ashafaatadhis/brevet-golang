package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository is a struct that represents a user repository
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindAllFiltered retrieves all users with optional filters, sorting, and pagination
func (r *UserRepository) FindAllFiltered(opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.User{}, &models.Profile{})
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

	db := r.db.Model(&models.User{})

	joinConditions := map[string]string{
		"profile": "LEFT JOIN profiles AS profiles ON profiles.user_id = users.id",
	}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "users", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("name ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var users []models.User
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("Profile").
		Find(&users).Error

	return users, total, err
}

// FindByID retrieves a user by their ID, including their profile
func (r *UserRepository) FindByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Save updates an existing user or creates a new one if it doesn't exist
func (r *UserRepository) Save(user *models.User) error {
	if user.Profile != nil {
		user.Profile.UserID = user.ID
	}
	return r.db.Save(user).Error
}

// DeleteByID deletes a user by their ID
func (r *UserRepository) DeleteByID(userID uuid.UUID) error {
	result := r.db.Delete(&models.User{}, "id = ?", userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// SaveUser saves a user within a transaction
func (r *UserRepository) SaveUser(tx *gorm.DB, user *models.User) error {
	if user.Profile != nil {
		user.Profile.UserID = user.ID
	}
	return tx.Save(user).Error
}

// DeleteUser deletes a user by their ID
func (r *UserRepository) DeleteUser(userID uuid.UUID) error {
	result := r.db.Delete(&models.User{}, "id = ?", userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
