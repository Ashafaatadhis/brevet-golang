package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttendanceRepository provides methods for managing assignments
type AttendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository creates a new assignment repository
func NewAttendanceRepository(db *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// WithTx running with transaction
func (r *AttendanceRepository) WithTx(tx *gorm.DB) *AttendanceRepository {
	return &AttendanceRepository{db: tx}
}

// GetAllFilteredAttendances retrieves all attendances with pagination and filtering options
func (r *AttendanceRepository) GetAllFilteredAttendances(opts utils.QueryOptions) ([]models.Attendance, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Attendance{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Preload("User").Model(&models.Attendance{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "attendances", opts.Filters, validSortFields, joinConditions, joinedRelations)

	var total int64
	db.Count(&total)

	var attendances []models.Attendance
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&attendances).Error

	return attendances, total, err
}

// GetAllFilteredAttendancesByBatchSlug retrieves all attendances with pagination and filtering options
func (r *AttendanceRepository) GetAllFilteredAttendancesByBatchSlug(batchSlug string, opts utils.QueryOptions) ([]models.Attendance, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Attendance{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Preload("User").Model(&models.Attendance{}).
		Joins("JOIN meetings ON meetings.id = attendances.meeting_id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("batches.slug = ?", batchSlug)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "attendances", opts.Filters, validSortFields, joinConditions, joinedRelations)

	var total int64
	db.Count(&total)

	var attendances []models.Attendance
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&attendances).Error

	return attendances, total, err
}

// Create for create new attendance
func (r *AttendanceRepository) Create(attendance *models.Attendance) error {
	return r.db.Create(attendance).Error
}

// FindByID for find attendance by id
func (r *AttendanceRepository) FindByID(id uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Preload("User").First(&attendance, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

// GetByMeetingAndUser method for get attendance by meeting and user id
func (r *AttendanceRepository) GetByMeetingAndUser(meetingID, userID uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&attendance).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

// UpdateByMeetingAndUser for update attending
func (r *AttendanceRepository) UpdateByMeetingAndUser(meetingID, userID uuid.UUID, update *models.Attendance) error {
	return r.db.
		Model(&models.Attendance{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Updates(map[string]any{
			"is_present": update.IsPresent,
			"note":       update.Note,
			"updated_by": update.UpdatedBy,
		}).Error
}
