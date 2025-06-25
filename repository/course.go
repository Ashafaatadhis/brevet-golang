package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CourseRepository is a struct that represents a course repository
type CourseRepository struct {
	db *gorm.DB
}

// NewCourseRepository creates a new course repository
func NewCourseRepository(db *gorm.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// GetAllFilteredCourses retrieves all courses with pagination and filtering options
func (r *CourseRepository) GetAllFilteredCourses(opts utils.QueryOptions) ([]models.Course, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Course{})
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

	db := r.db.Model(&models.Course{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "courses", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var courses []models.Course
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("CourseImages").
		Find(&courses).Error

	return courses, total, err
}

// GetCourseBySlug retrieves a course by its slug
func (r *CourseRepository) GetCourseBySlug(slug string) (*models.Course, error) {
	var course models.Course
	err := r.db.Preload("CourseImages").First(&course, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// CreateTx inserts a new course within a transaction
func (r *CourseRepository) CreateTx(db *gorm.DB, course *models.Course) error {
	return db.Create(course).Error
}

// CreateCourseImagesBulkTx inserts multiple course images into the database
func (r *CourseRepository) CreateCourseImagesBulkTx(tx *gorm.DB, images []models.CourseImage) error {
	if len(images) == 0 {
		return nil
	}
	return tx.Create(&images).Error
}

// FindByIDWithImages retrieves a course by its ID along with its associated images
func (r *CourseRepository) FindByIDWithImages(id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := r.db.Preload("CourseImages").First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// FindByIDWithImagesTx retrieves a course by its ID along with its associated images within a transaction
func (r *CourseRepository) FindByIDWithImagesTx(db *gorm.DB, id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := db.Preload("CourseImages").First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// IsSlugExists checks if a course slug already exists in the database
func (r *CourseRepository) IsSlugExists(slug string) bool {
	var count int64
	r.db.Model(&models.Course{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

// UpdateTx updates an existing course within a transaction
func (r *CourseRepository) UpdateTx(tx *gorm.DB, course *models.Course) error {
	return tx.Save(course).Error
}

// DeleteCourseImagesByCourseID deletes all images associated with a course by its ID
func (r *CourseRepository) DeleteCourseImagesByCourseID(tx *gorm.DB, courseID uuid.UUID) error {
	return tx.Where("course_id = ?", courseID).Delete(&models.CourseImage{}).Error
}

// FindByID retrieves a course by its ID
func (r *CourseRepository) FindByID(db *gorm.DB, id uuid.UUID) (*models.Course, error) {
	var course models.Course
	err := db.First(&course, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// DeleteByIDTx deletes a course by its ID within a transaction
func (r *CourseRepository) DeleteByIDTx(tx *gorm.DB, id uuid.UUID) error {
	return tx.Where("id = ?", id).Delete(&models.Course{}).Error
}
