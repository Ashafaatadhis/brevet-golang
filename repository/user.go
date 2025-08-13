package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IUserRepository interfacee
type IUserRepository interface {
	FindAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.User, int64, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	FindByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error)
	Create(ctx context.Context, user *models.User) error
	CreateProfile(ctx context.Context, profile *models.Profile) error
	Save(ctx context.Context, user *models.User) error
	DeleteByID(ctx context.Context, userID uuid.UUID) error
	SaveProfile(ctx context.Context, profile *models.Profile) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

// IUserTXRepository interface
type IUserTXRepository interface {
	IUserRepository
	WithTx(tx *gorm.DB) IUserRepository
}

// UserRepository is a struct that represents a user repository
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) IUserTXRepository {
	return &UserRepository{db: db}
}

// WithTx running with transaction
func (r *UserRepository) WithTx(tx *gorm.DB) IUserRepository {
	return &UserRepository{db: tx}
}

// FindAllFiltered retrieves all users with optional filters, sorting, and pagination
func (r *UserRepository) FindAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.User, int64, error) {
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

	db := r.db.WithContext(ctx).Model(&models.User{})

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
func (r *UserRepository) FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDs is get all users in ids
func (r *UserRepository) FindByIDs(ctx context.Context, userIDs []uuid.UUID) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// CreateProfile saves a user profile within a transaction
func (r *UserRepository) CreateProfile(ctx context.Context, profile *models.Profile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

// Save updates an existing user or creates a new one if it doesn't exist
func (r *UserRepository) Save(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		// Tangani error duplicate phone_number dari Postgres
		if strings.Contains(err.Error(), "duplicate key") &&
			strings.Contains(err.Error(), "phone_number") {
			return fmt.Errorf("nomor telepon sudah digunakan")
		}
	}

	return err
}

// DeleteByID deletes a user by their ID
func (r *UserRepository) DeleteByID(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// SaveProfile saves a user profile within a transaction
func (r *UserRepository) SaveProfile(ctx context.Context, profile *models.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

// DeleteUser deletes a user by their ID
func (r *UserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
