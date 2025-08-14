package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IMaterialRepository interface
type IMaterialRepository interface {
	WithTx(tx *gorm.DB) IMaterialRepository
	GetAllFilteredMaterial(ctx context.Context, opts utils.QueryOptions) ([]models.Material, int64, error)
	GetAllFilteredMaterialsByMeetingID(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Material, int64, error)
	Create(ctx context.Context, assignment *models.Material) error
	Update(ctx context.Context, assignment *models.Material) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Material, error)
}

// MaterialRepository is init struct
type MaterialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository creates a new material repository
func NewMaterialRepository(db *gorm.DB) IMaterialRepository {
	return &MaterialRepository{db: db}
}

// WithTx running with transaction
func (r *MaterialRepository) WithTx(tx *gorm.DB) IMaterialRepository {
	return &MaterialRepository{db: tx}
}

// GetAllFilteredMaterial retrieves all Material with pagination and filtering options
func (r *MaterialRepository) GetAllFilteredMaterial(ctx context.Context, opts utils.QueryOptions) ([]models.Material, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Material{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Material{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "materials", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Joins("LEFT JOIN meetings ON meetings.id = materials.meeting_id")
		db = db.Where("materials.title ILIKE ? OR meetings.title ILIKE ?", "%"+opts.Search+"%", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var materials []models.Material
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&materials).Error

	return materials, total, err
}

// GetAllFilteredMaterialsByMeetingID retrieves all materials with pagination and filtering options
func (r *MaterialRepository) GetAllFilteredMaterialsByMeetingID(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Material, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Material{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Material{}).
		Where("meeting_id = ?", meetingID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "materials", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Joins("LEFT JOIN meetings ON meetings.id = materials.meeting_id")
		db = db.Where("materials.title ILIKE ? OR meetings.title ILIKE ?", "%"+opts.Search+"%", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var materials []models.Material
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&materials).Error

	return materials, total, err
}

// Create creates a new material
func (r *MaterialRepository) Create(ctx context.Context, assignment *models.Material) error {
	return r.db.WithContext(ctx).Create(assignment).Error
}

// Update updates an existing material
func (r *MaterialRepository) Update(ctx context.Context, assignment *models.Material) error {
	return r.db.WithContext(ctx).Save(assignment).Error
}

// DeleteByID deletes an material by its ID
func (r *MaterialRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Material{}).Error
}

// FindByID retrieves a meeting by its ID
func (r *MaterialRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Material, error) {
	var material models.Material
	err := r.db.WithContext(ctx).First(&material, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}
