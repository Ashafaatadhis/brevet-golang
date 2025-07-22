package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BlogRepository is a struct that represents a blog repository
type BlogRepository struct {
	db *gorm.DB
}

// NewBlogRepository creates a new blog repository
func NewBlogRepository(db *gorm.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// WithTx running with transaction
func (r *BlogRepository) WithTx(tx *gorm.DB) *BlogRepository {
	return &BlogRepository{db: tx}
}

// GetAllFilteredBlogs retrieves all blogs with pagination and filtering options
func (r *BlogRepository) GetAllFilteredBlogs(opts utils.QueryOptions) ([]models.Blog, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Blog{})
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

	db := r.db.Model(&models.Blog{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "blogs", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var blogs []models.Blog
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&blogs).Error

	return blogs, total, err
}

// GetBlogBySlug retrieves a blog by its slug
func (r *BlogRepository) GetBlogBySlug(slug string) (*models.Blog, error) {
	var blog models.Blog
	err := r.db.First(&blog, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// IsSlugExists checks if a course slug already exists in the database
func (r *BlogRepository) IsSlugExists(slug string) bool {
	var count int64
	r.db.Model(&models.Blog{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// Create creates a new blog in the database
func (r *BlogRepository) Create(blog *models.Blog) error {
	return r.db.Create(blog).Error
}

// FindByID retrieves a blog by its ID
func (r *BlogRepository) FindByID(id uuid.UUID) (*models.Blog, error) {
	var blog models.Blog
	err := r.db.First(&blog, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// Update updates an existing blog in the database
func (r *BlogRepository) Update(blog *models.Blog) error {
	return r.db.Save(blog).Error
}

// DeleteByID deletes a blog by its ID
func (r *BlogRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Blog{}).Error
}
