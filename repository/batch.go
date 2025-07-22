package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchRepository is a struct that represents a batch repository
type BatchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a new batch repository
func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

// WithTx running with transaction
func (r *BatchRepository) WithTx(tx *gorm.DB) *BatchRepository {
	return &BatchRepository{db: tx}
}

// WithLock running with transaction and lock
func (r *BatchRepository) WithLock() *BatchRepository {
	return &BatchRepository{
		db: r.db.Clauses(clause.Locking{Strength: "UPDATE"}),
	}
}

// GetAllFilteredBatches retrieves all batches with pagination and filtering options
func (r *BatchRepository) GetAllFilteredBatches(opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Batch{}, &models.BatchDay{})
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

	db := r.db.Preload("BatchDays").Model(&models.Batch{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batch []models.Batch
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetAllFilteredBatchesByCourseSlug retrieves all filtered batches by course slug
func (r *BatchRepository) GetAllFilteredBatchesByCourseSlug(courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Batch{}, &models.BatchDay{})
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

	db := r.db.Preload("BatchDays").Model(&models.Batch{}).Where("course_id = ?", courseID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var batch []models.Batch
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetBatchBySlug retrieves a batch by its slug
func (r *BatchRepository) GetBatchBySlug(slug string) (*models.Batch, error) {
	var batch models.Batch
	err := r.db.Preload("BatchDays").First(&batch, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// IsSlugExists checks if a batch slug already exists in the database
func (r *BatchRepository) IsSlugExists(slug string) bool {
	var count int64
	r.db.Model(&models.Batch{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// Create inserts a new batch
func (r *BatchRepository) Create(batch *models.Batch) error {
	return r.db.Create(batch).Error
}

// Update updates an existing batch
func (r *BatchRepository) Update(batch *models.Batch) error {
	return r.db.Save(batch).Error
}

// FindByID retrieves a batch by its ID
func (r *BatchRepository) FindByID(id uuid.UUID) (*models.Batch, error) {
	var batch models.Batch
	err := r.db.Preload("BatchDays").First(&batch, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// DeleteByID deletes a batch by its ID
func (r *BatchRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Batch{}).Error
}

// GetAllTeacherInBatch get all teacher in batch
func (r *BatchRepository) GetAllTeacherInBatch(batchID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.User{})
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

	db := r.db.
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
	err = db.
		Order(fmt.Sprintf("users.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&users).Error

	return users, total, err
}

// GetBatchesByUserPurchaseFiltered is repository for get all batches where user purchase
func (r *BatchRepository) GetBatchesByUserPurchaseFiltered(userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Batch{}, &models.BatchDay{})
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

	// JOIN ke purchases, dan preload relasi
	db := r.db.
		Joins("JOIN purchases ON purchases.batch_id = batches.id").
		Preload("BatchDays").
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
	err = db.
		Order(fmt.Sprintf("batches.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batch).Error

	return batch, total, err
}

// GetBatchesByGuruMeetingRelationFiltered is repo for get all batches where has taught
func (r *BatchRepository) GetBatchesByGuruMeetingRelationFiltered(guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Batch{}, &models.BatchDay{}, &models.User{})
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

	db := r.db.
		Joins("JOIN meetings ON meetings.batch_id = batches.id").
		Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
		Preload("BatchDays").
		Model(&models.Batch{}).
		Where("meeting_teacher.user_id = ?", guruID).
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
	err = db.
		Order(fmt.Sprintf("batches.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&batches).Error

	return batches, total, err
}
