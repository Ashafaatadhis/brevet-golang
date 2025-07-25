package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignmentRepository provides methods for managing assignments
type AssignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// WithTx running with transaction
func (r *AssignmentRepository) WithTx(tx *gorm.DB) *AssignmentRepository {
	return &AssignmentRepository{db: tx}
}

// GetAllFilteredAssignments retrieves all assignments with pagination and filtering options
func (r *AssignmentRepository) GetAllFilteredAssignments(opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Assignment{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Preload("AssignmentFiles").Model(&models.Assignment{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "assignments", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var assignments []models.Assignment
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&assignments).Error

	return assignments, total, err
}

// Create creates a new assignment
func (r *AssignmentRepository) Create(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

// CreateFiles creates new assignment files
func (r *AssignmentRepository) CreateFiles(assignmentFiles []models.AssignmentFiles) error {
	return r.db.Create(assignmentFiles).Error
}

// FindByID retrieves a meeting by its ID
func (r *AssignmentRepository) FindByID(id uuid.UUID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.Preload("AssignmentFiles").First(&assignment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}
