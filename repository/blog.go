package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IBlogRepository interface
type IBlogRepository interface {
	WithTx(tx *gorm.DB) IBlogRepository
	GetAllFilteredBlogs(ctx context.Context, opts utils.QueryOptions) ([]models.Blog, int64, error)
	GetBlogBySlug(ctx context.Context, slug string) (*models.Blog, error)
	IsSlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, blog *models.Blog) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Blog, error)
	Update(ctx context.Context, blog *models.Blog) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

// BlogRepository is a struct that represents a blog repository
type BlogRepository struct {
	db *gorm.DB
}

// NewBlogRepository creates a new blog repository
func NewBlogRepository(db *gorm.DB) IBlogRepository {
	return &BlogRepository{db: db}
}

// WithTx running with transaction
func (r *BlogRepository) WithTx(tx *gorm.DB) IBlogRepository {
	return &BlogRepository{db: tx}
}

// GetAllFilteredBlogs retrieves all blogs with pagination and filtering options
func (r *BlogRepository) GetAllFilteredBlogs(ctx context.Context, opts utils.QueryOptions) ([]models.Blog, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Blog{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Blog{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "blogs", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var blogs []models.Blog
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&blogs).Error

	return blogs, total, err
}

// GetBlogBySlug retrieves a blog by its slug
func (r *BlogRepository) GetBlogBySlug(ctx context.Context, slug string) (*models.Blog, error) {
	var blog models.Blog
	err := r.db.WithContext(ctx).First(&blog, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// IsSlugExists checks if a course slug already exists in the database
func (r *BlogRepository) IsSlugExists(ctx context.Context, slug string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&models.Blog{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// Create creates a new blog in the database
func (r *BlogRepository) Create(ctx context.Context, blog *models.Blog) error {
	return r.db.WithContext(ctx).Create(blog).Error
}

// FindByID retrieves a blog by its ID
func (r *BlogRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Blog, error) {
	var blog models.Blog
	err := r.db.WithContext(ctx).First(&blog, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// Update updates an existing blog in the database
func (r *BlogRepository) Update(ctx context.Context, blog *models.Blog) error {
	return r.db.WithContext(ctx).Save(blog).Error
}

// DeleteByID deletes a blog by its ID
func (r *BlogRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Blog{}).Error
}
