package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ICourseRepository interface
type ICourseRepository interface {
	WithTx(tx *gorm.DB) ICourseRepository
	GetAllFilteredCourses(ctx context.Context, opts utils.QueryOptions) ([]models.Course, int64, error)
	GetCourseBySlug(ctx context.Context, slug string) (*models.Course, error)
	Create(ctx context.Context, course *models.Course) error
	CreateCourseImagesBulk(ctx context.Context, images []models.CourseImage) error
	FindByIDWithImages(ctx context.Context, id uuid.UUID) (*models.Course, error)
	IsSlugExists(ctx context.Context, slug string) bool
	Update(ctx context.Context, course *models.Course) error
	DeleteCourseImagesByCourseID(ctx context.Context, courseID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Course, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

// CourseRepository is a struct that represents a course repository
type CourseRepository struct {
	db *gorm.DB
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(db *gorm.DB) ICourseRepository {
	return &CourseRepository{db: db}
}

// WithTx running with transaction
func (r *CourseRepository) WithTx(tx *gorm.DB) ICourseRepository {
	return &CourseRepository{db: tx}
}

// GetAllFilteredCourses retrieves all courses with pagination and filtering options
func (r *CourseRepository) GetAllFilteredCourses(ctx context.Context, opts utils.QueryOptions) ([]models.Course, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Course{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Course{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "courses", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var courses []models.Course
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("CourseImages").
		Find(&courses).Error

	return courses, total, err
}

// GetCourseBySlug retrieves a course by its slug
func (r *CourseRepository) GetCourseBySlug(ctx context.Context, slug string) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).Preload("CourseImages").First(&course, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// Create inserts a new course into the database
func (r *CourseRepository) Create(ctx context.Context, course *models.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

// CreateCourseImagesBulk inserts multiple course images into the database
func (r *CourseRepository) CreateCourseImagesBulk(ctx context.Context, images []models.CourseImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&images).Error
}

// FindByIDWithImages retrieves a course by its ID along with its associated images
func (r *CourseRepository) FindByIDWithImages(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).Preload("CourseImages").First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// IsSlugExists checks if a course slug already exists in the database
func (r *CourseRepository) IsSlugExists(ctx context.Context, slug string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&models.Course{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// Update updates an existing course
func (r *CourseRepository) Update(ctx context.Context, course *models.Course) error {
	return r.db.WithContext(ctx).Save(course).Error
}

// DeleteCourseImagesByCourseID deletes all images associated with a course by its ID
func (r *CourseRepository) DeleteCourseImagesByCourseID(ctx context.Context, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("course_id = ?", courseID).Delete(&models.CourseImage{}).Error
}

// FindByID retrieves a course by its ID
func (r *CourseRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// DeleteByID deletes a course by its ID
func (r *CourseRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Course{}).Error
}
