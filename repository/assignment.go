package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAssignmentRepository interface
type IAssignmentRepository interface {
	WithTx(tx *gorm.DB) IAssignmentRepository
	GetAllFilteredAssignments(ctx context.Context, opts utils.QueryOptions) ([]models.Assignment, int64, error)
	GetAllFilteredAssignmentsByMeetingID(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Assignment, int64, error)
	Create(ctx context.Context, assignment *models.Assignment) error
	Update(ctx context.Context, assignment *models.Assignment) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	CreateFiles(ctx context.Context, assignmentFiles []models.AssignmentFiles) error
	DeleteFilesByAssignmentID(ctx context.Context, assignmentID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Assignment, error)
	GetBatchIDByAssignmentID(ctx context.Context, assignmentID uuid.UUID) (uuid.UUID, error)
	GetBatchByAssignmentID(ctx context.Context, assignmentID uuid.UUID) (models.Batch, error)
	CountByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error)
}

// AssignmentRepository provides methods for managing assignments
type AssignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *gorm.DB) IAssignmentRepository {
	return &AssignmentRepository{db: db}
}

// WithTx running with transaction
func (r *AssignmentRepository) WithTx(tx *gorm.DB) IAssignmentRepository {
	return &AssignmentRepository{db: tx}
}

// GetAllFilteredAssignments retrieves all assignments with pagination and filtering options
func (r *AssignmentRepository) GetAllFilteredAssignments(ctx context.Context, opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Assignment{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("AssignmentFiles").Model(&models.Assignment{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "assignments", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		// Join ke meetings agar bisa search by meetings.title
		db = db.Joins("LEFT JOIN meetings ON meetings.id = assignments.meeting_id")
		db = db.Where("assignments.title ILIKE ? OR meetings.title ILIKE ?", "%"+opts.Search+"%", "%"+opts.Search+"%")
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

// GetAllFilteredAssignmentsByMeetingID retrieves all assignments with pagination and filtering options
func (r *AssignmentRepository) GetAllFilteredAssignmentsByMeetingID(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Assignment{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("AssignmentFiles").Model(&models.Assignment{}).
		Where("meeting_id = ?", meetingID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "assignments", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		// Join ke meetings agar bisa search by meetings.title
		db = db.Joins("LEFT JOIN meetings ON meetings.id = assignments.meeting_id")
		db = db.Where("assignments.title ILIKE ? OR meetings.title ILIKE ?", "%"+opts.Search+"%", "%"+opts.Search+"%")
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
func (r *AssignmentRepository) Create(ctx context.Context, assignment *models.Assignment) error {
	return r.db.WithContext(ctx).Create(assignment).Error
}

// Update updates an existing assignment
func (r *AssignmentRepository) Update(ctx context.Context, assignment *models.Assignment) error {
	return r.db.WithContext(ctx).Save(assignment).Error
}

// DeleteByID deletes an assignment by its ID
func (r *AssignmentRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Preload("AssignmentFiles").Where("id = ?", id).Delete(&models.Assignment{}).Error
}

// CreateFiles creates new assignment files
func (r *AssignmentRepository) CreateFiles(ctx context.Context, assignmentFiles []models.AssignmentFiles) error {
	return r.db.WithContext(ctx).Create(assignmentFiles).Error
}

// DeleteFilesByAssignmentID deletes all files associated with a specific assignment
func (r *AssignmentRepository) DeleteFilesByAssignmentID(ctx context.Context, assignmentID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("assignment_id = ?", assignmentID).Delete(&models.AssignmentFiles{}).Error
}

// FindByID retrieves a meeting by its ID
func (r *AssignmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.WithContext(ctx).Preload("AssignmentFiles").First(&assignment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetBatchIDByAssignmentID get batch
func (r *AssignmentRepository) GetBatchIDByAssignmentID(ctx context.Context, assignmentID uuid.UUID) (uuid.UUID, error) {
	var assignment models.Assignment
	err := r.db.WithContext(ctx).Preload("Meeting").
		First(&assignment, "id = ?", assignmentID).Error
	if err != nil {
		return uuid.Nil, err
	}

	return assignment.Meeting.BatchID, nil
}

// GetBatchByAssignmentID ambil batch dari assignment
func (r *AssignmentRepository) GetBatchByAssignmentID(ctx context.Context, assignmentID uuid.UUID) (models.Batch, error) {
	var assignment models.Assignment
	err := r.db.WithContext(ctx).
		Preload("Meeting.Batch").
		First(&assignment, "id = ?", assignmentID).Error
	if err != nil {
		return models.Batch{}, err
	}

	return assignment.Meeting.Batch, nil
}

// CountByBatchID for count assignment by batch id
func (r *AssignmentRepository) CountByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Assignment{}).
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Where("meetings.batch_id = ?", batchID).
		Count(&count).Error
	return count, err
}
