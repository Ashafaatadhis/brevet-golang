package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IBatchRepository interface
type IBatchRepository interface {
	WithTx(tx *gorm.DB) IBatchRepository
	WithLock() IBatchRepository
	GetAllFilteredBatches(ctx context.Context, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetAllFilteredBatchesByCourseSlug(ctx context.Context, courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetBatchBySlug(ctx context.Context, slug string) (*models.Batch, error)
	IsSlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, batch *models.Batch) error
	Update(ctx context.Context, batch *models.Batch) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Batch, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetAllTeacherInBatch(ctx context.Context, batchID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error)
	GetBatchesByUserPurchaseFiltered(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetBatchesByGuruMeetingRelationFiltered(ctx context.Context, guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
}

// BatchRepository is a struct that represents a batch repository
type BatchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a new batch repository
func NewBatchRepository(db *gorm.DB) IBatchRepository {
	return &BatchRepository{db: db}
}

// WithTx running with transaction
func (r *BatchRepository) WithTx(tx *gorm.DB) IBatchRepository {
	return &BatchRepository{db: tx}
}

// WithLock running with transaction and lock
func (r *BatchRepository) WithLock() IBatchRepository {
	return &BatchRepository{
		db: r.db.Clauses(clause.Locking{Strength: "UPDATE"}),
	}
}

// GetAllFilteredBatches retrieves all batches with pagination and filtering options
func (r *BatchRepository) GetAllFilteredBatches(ctx context.Context, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Batch{}, &models.BatchDay{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("BatchDays").Preload("BatchGroups").Model(&models.Batch{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batch []models.Batch
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetAllFilteredBatchesByCourseSlug retrieves all filtered batches by course slug
func (r *BatchRepository) GetAllFilteredBatchesByCourseSlug(ctx context.Context, courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Batch{}, &models.BatchDay{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("BatchDays").Preload("BatchGroups").Model(&models.Batch{}).Where("course_id = ?", courseID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batch []models.Batch
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetBatchBySlug retrieves a batch by its slug
func (r *BatchRepository) GetBatchBySlug(ctx context.Context, slug string) (*models.Batch, error) {
	var batch models.Batch
	err := r.db.WithContext(ctx).Preload("BatchDays").Preload("BatchGroups").First(&batch, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// IsSlugExists checks if a batch slug already exists in the database
func (r *BatchRepository) IsSlugExists(ctx context.Context, slug string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&models.Batch{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// Create inserts a new batch
func (r *BatchRepository) Create(ctx context.Context, batch *models.Batch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

// Update updates an existing batch
func (r *BatchRepository) Update(ctx context.Context, batch *models.Batch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}

// FindByID retrieves a batch by its ID
func (r *BatchRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Batch, error) {
	var batch models.Batch
	err := r.db.WithContext(ctx).Preload("BatchDays").Preload("BatchGroups").First(&batch, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// DeleteByID deletes a batch by its ID
func (r *BatchRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Batch{}).Error
}

// GetAllTeacherInBatch get all teacher in batch
func (r *BatchRepository) GetAllTeacherInBatch(ctx context.Context, batchID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.User{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).
		Model(&models.User{}).
		Joins("JOIN batch_teachers bt ON bt.user_id = users.id").
		Where("bt.batch_id = ?", batchID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "users", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// Optional: filter dan search
	if opts.Search != "" {
		db = db.Where("users.name ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	err := db.
		Order(fmt.Sprintf("users.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&users).Error

	return users, total, err
}

// GetBatchesByUserPurchaseFiltered is repository for get all batches where user purchase
func (r *BatchRepository) GetBatchesByUserPurchaseFiltered(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Batch{}, &models.BatchDay{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// JOIN ke purchases, dan preload relasi
	db := r.db.WithContext(ctx).
		Joins("JOIN purchases ON purchases.batch_id = batches.id").
		Preload("BatchDays").
		Preload("BatchGroups").
		Model(&models.Batch{}).
		Where("purchases.user_id = ?", userID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	// Apply dynamic filter (dari query param)
	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// Search by title (opsional)
	if opts.Search != "" {
		db = db.Where("batches.title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batch []models.Batch
	err := db.
		Order(fmt.Sprintf("batches.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetBatchesByGuruMeetingRelationFiltered is repo for get all batches where has taught
func (r *BatchRepository) GetBatchesByGuruMeetingRelationFiltered(ctx context.Context, guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Batch{}, &models.BatchDay{}, &models.User{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).
		Joins("JOIN meetings ON meetings.batch_id = batches.id").
		Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
		Preload("BatchDays").
		Preload("BatchGroups").
		Model(&models.Batch{}).
		Where("meeting_teachers.user_id = ?", guruID).
		Group("batches.id")

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("batches.title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batches []models.Batch
	err := db.
		Order(fmt.Sprintf("batches.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batches).Error

	return batches, total, err
}
