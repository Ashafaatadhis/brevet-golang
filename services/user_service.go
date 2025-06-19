package services

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

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

// GetUserByID is a method that returns a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveUser untuk update user
func (s *UserService) SaveUser(user *models.User) error {
	// Jika Profile belum nil, pastikan foreign key-nya benar
	if user.Profile != nil {
		user.Profile.UserID = user.ID
	}

	// Simpan user (beserta profile)
	if err := s.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

// DeleteUser is for delete user
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	result := s.db.Delete(&models.User{}, "id = ?", userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
