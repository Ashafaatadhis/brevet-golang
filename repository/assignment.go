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
func (r *AssignmentRepository) GetAllFilteredAssignmentsByMeetingID(meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Assignment{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Preload("AssignmentFiles").Model(&models.Assignment{}).
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

// IsUserTeachingInMeeting for know user is teacher in this meet
func (r *MeetingRepository) IsUserTeachingInMeeting(userID, meetingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("meeting_teachers").
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create creates a new assignment
func (r *AssignmentRepository) Create(assignment *models.Assignment) error {
	return r.db.Create(assignment).Error
}

// Update updates an existing assignment
func (r *AssignmentRepository) Update(assignment *models.Assignment) error {
	return r.db.Save(assignment).Error
}

// DeleteByID deletes an assignment by its ID
func (r *AssignmentRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Preload("AssignmentFiles").Where("id = ?", id).Delete(&models.Assignment{}).Error
}

// CreateFiles creates new assignment files
func (r *AssignmentRepository) CreateFiles(assignmentFiles []models.AssignmentFiles) error {
	return r.db.Create(assignmentFiles).Error
}

// DeleteFilesByAssignmentID deletes all files associated with a specific assignment
func (r *AssignmentRepository) DeleteFilesByAssignmentID(assignmentID uuid.UUID) error {
	return r.db.Where("assignment_id = ?", assignmentID).Delete(&models.AssignmentFiles{}).Error
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
