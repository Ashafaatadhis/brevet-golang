package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ITestimonialRepository interface
type ITestimonialRepository interface {
	WithTx(tx *gorm.DB) ITestimonialRepository
	GetByUserAndBatch(ctx context.Context, userID, batchID uuid.UUID) (*models.Testimonial, error)
	GetAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.Testimonial, int64, error)
	GetByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Testimonial, int64, error)
	Create(ctx context.Context, testimonial *models.Testimonial) error
	Update(ctx context.Context, testimonial *models.Testimonial) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Testimonial, error)
}

// TestimonialRepository repository
type TestimonialRepository struct {
	db *gorm.DB
}

// NewTestimonialRepository init
func NewTestimonialRepository(db *gorm.DB) ITestimonialRepository {
	return &TestimonialRepository{db: db}
}

// WithTx running with transaction
func (r *TestimonialRepository) WithTx(tx *gorm.DB) ITestimonialRepository {
	return &TestimonialRepository{db: tx}
}

// GetByUserAndBatch mengambil testimonial berdasarkan userID dan batchID
func (r *TestimonialRepository) GetByUserAndBatch(ctx context.Context, userID, batchID uuid.UUID) (*models.Testimonial, error) {
	var testimonial models.Testimonial

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND batch_id = ?", userID, batchID).
		First(&testimonial).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &testimonial, nil
}

// GetByBatchSlugFiltered Get all testimonials by batch_slug (with pagination/filter)
func (r *TestimonialRepository) GetByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Testimonial, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Testimonial{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).
		Model(&models.Testimonial{}).
		Joins("JOIN batches b ON testimonials.batch_id = b.id").
		Where("b.slug = ?", batchSlug).
		Preload("User").
		Preload("Batch")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var testimonials []models.Testimonial
	err := db.Order(sort + " " + order).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&testimonials).Error

	return testimonials, total, err
}

// GetAllFiltered Get all testimonials (with pagination/filter)
func (r *TestimonialRepository) GetAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.Testimonial, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Testimonial{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).
		Model(&models.Testimonial{}).
		Preload("User").
		Preload("Batch")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var testimonials []models.Testimonial
	err := db.Order(sort + " " + order).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&testimonials).Error

	return testimonials, total, err
}

// Create testimonial
func (r *TestimonialRepository) Create(ctx context.Context, testimonial *models.Testimonial) error {
	return r.db.WithContext(ctx).Create(testimonial).Error
}

// Update testimonial
func (r *TestimonialRepository) Update(ctx context.Context, testimonial *models.Testimonial) error {
	return r.db.WithContext(ctx).Save(testimonial).Error
}

// Delete by id testimonial
func (r *TestimonialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Testimonial{}, "id = ?", id).Error
}

// GetByID get by id testimonial
func (r *TestimonialRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Testimonial, error) {
	var testimonial models.Testimonial
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Batch").
		First(&testimonial, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &testimonial, nil
}
