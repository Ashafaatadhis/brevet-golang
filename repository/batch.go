package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BatchRepository is a struct that represents a batch repository
type BatchRepository struct {
	db *gorm.DB
}

// NewBatchRepository creates a new batch repository
func NewBatchRepository(db *gorm.DB) *BatchRepository {
	return &BatchRepository{db: db}
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

// CreateTx inserts a new batch within a transaction
func (r *BatchRepository) CreateTx(db *gorm.DB, batch *models.Batch) error {
	return db.Create(batch).Error
}

// CreateBatchTeacherTx inserts a new batch with teacher information within a transaction
func (r *BatchRepository) CreateBatchTeacherTx(db *gorm.DB, batch *models.BatchTeacher) error {
	return db.Create(batch).Error
}

// UpdateTx updates an existing batch within a transaction
func (r *BatchRepository) UpdateTx(tx *gorm.DB, batch *models.Batch) error {
	return tx.Save(batch).Error
}

// FindByIDTx retrieves a batch by its ID
func (r *BatchRepository) FindByIDTx(db *gorm.DB, id uuid.UUID) (*models.Batch, error) {
	var batch models.Batch
	err := db.First(&batch, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
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

// DeleteByIDTx deletes a batch by its ID within a transaction
func (r *BatchRepository) DeleteByIDTx(tx *gorm.DB, id uuid.UUID) error {
	return tx.Where("id = ?", id).Delete(&models.Batch{}).Error
}

// IsTeacherAssigned checks whether a user is already assigned as a teacher in a batch
func (r *BatchRepository) IsTeacherAssigned(batchID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.BatchTeacher{}).
		Where("batch_id = ? AND user_id = ?", batchID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
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

// FindBatchTeacherByBatchIDAndUserIDTx this function for Find batchTeacher where userid and batchid with TX (Transaction)
func (r *BatchRepository) FindBatchTeacherByBatchIDAndUserIDTx(db *gorm.DB, batchID uuid.UUID, userID uuid.UUID) (*models.BatchTeacher, error) {
	var batchTeacher models.BatchTeacher
	err := db.First(&batchTeacher, "batch_id = ? AND user_id = ?", batchID, userID).Error
	if err != nil {
		return nil, err
	}
	return &batchTeacher, nil
}

// DeleteTeacherByIDTx delete teacher from batchTeacher
func (r *BatchRepository) DeleteTeacherByIDTx(tx *gorm.DB, batchID uuid.UUID, userID uuid.UUID) error {
	return tx.Where("batch_id = ? AND user_id = ?", batchID, userID).Delete(&models.BatchTeacher{}).Error
}
