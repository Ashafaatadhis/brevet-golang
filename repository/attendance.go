package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IAttendanceRepository interface
type IAttendanceRepository interface {
	WithTx(tx *gorm.DB) IAttendanceRepository
	GetAllFilteredAttendances(ctx context.Context, opts utils.QueryOptions) ([]models.Attendance, int64, error)
	// GetAllFilteredAttendancesByBatchSlug(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Attendance, int64, error)
	GetAllFilteredAttendancesByBatchSlug(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error)
	Create(ctx context.Context, attendance *models.Attendance) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error)
	GetByMeetingAndUser(ctx context.Context, meetingID, userID uuid.UUID) (*models.Attendance, error)
	UpdateByMeetingAndUser(ctx context.Context, meetingID, userID uuid.UUID, update *models.Attendance) error
	CountByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error)
}

// AttendanceRepository provides methods for managing assignments
type AttendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository creates a new assignment repository
func NewAttendanceRepository(db *gorm.DB) IAttendanceRepository {
	return &AttendanceRepository{db: db}
}

// WithTx running with transaction
func (r *AttendanceRepository) WithTx(tx *gorm.DB) IAttendanceRepository {
	return &AttendanceRepository{db: tx}
}

// GetAllFilteredAttendances retrieves all attendances with pagination and filtering options
func (r *AttendanceRepository) GetAllFilteredAttendances(ctx context.Context, opts utils.QueryOptions) ([]models.Attendance, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Attendance{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("User").Model(&models.Attendance{})

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

// // GetAllFilteredAttendancesByBatchSlug retrieves all attendances with pagination and filtering options
// func (r *AttendanceRepository) GetAllFilteredAttendancesByBatchSlug(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Attendance, int64, error) {
// 	validSortFields := utils.GetValidColumnsFromStruct(&models.Attendance{})

// 	sort := opts.Sort
// 	if !validSortFields[sort] {
// 		sort = "id"
// 	}

// 	order := opts.Order
// 	if order != "asc" && order != "desc" {
// 		order = "asc"
// 	}

// 	db := r.db.WithContext(ctx).Preload("User").Model(&models.Attendance{}).
// 		Joins("JOIN meetings ON meetings.id = attendances.meeting_id").
// 		Joins("JOIN batches ON batches.id = meetings.batch_id").
// 		Where("batches.slug = ?", batchSlug)

// 	joinConditions := map[string]string{}
// 	joinedRelations := map[string]bool{}

// 	db = utils.ApplyFiltersWithJoins(db, "attendances", opts.Filters, validSortFields, joinConditions, joinedRelations)

// 	var total int64
// 	db.Count(&total)

// 	var attendances []models.Attendance
// 	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
// 		Limit(opts.Limit).
// 		Offset(opts.Offset).
// 		Find(&attendances).Error

// 	return attendances, total, err
// }

// GetAllFilteredAttendancesByBatchSlug retrieves all students in a batch with their attendance status
func (r *AttendanceRepository) GetAllFilteredAttendancesByBatchSlug(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.User{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Ambil meeting IDs dari batch slug
	var meetingIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&models.Meeting{}).
		Select("meetings.id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("batches.slug = ?", batchSlug).
		Find(&meetingIDs).Error; err != nil {
		return nil, 0, err
	}

	db := r.db.WithContext(ctx).
		Model(&models.User{}).
		Joins("JOIN purchases ON purchases.user_id = users.id").
		Joins("JOIN batches ON batches.id = purchases.batch_id").
		Where("batches.slug = ?", batchSlug).
		Where("users.role_type = ?", models.RoleTypeSiswa)

	// Apply filters & search
	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}
	db = utils.ApplyFiltersWithJoins(db, "users", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		q := "%" + opts.Search + "%"
		db = db.Where("users.name ILIKE ? OR users.email ILIKE ?", q, q)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []models.User
	err := db.
		Preload("Attendances", "meeting_id IN ?", meetingIDs). // preload attendances by meeting IDs
		Order(fmt.Sprintf("users.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// Create for create new attendance
func (r *AttendanceRepository) Create(ctx context.Context, attendance *models.Attendance) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

// FindByID for find attendance by id
func (r *AttendanceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.WithContext(ctx).Preload("User").First(&attendance, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

// GetByMeetingAndUser method for get attendance by meeting and user id
func (r *AttendanceRepository) GetByMeetingAndUser(ctx context.Context, meetingID, userID uuid.UUID) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.WithContext(ctx).Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&attendance).Error
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

// UpdateByMeetingAndUser for update attending
func (r *AttendanceRepository) UpdateByMeetingAndUser(ctx context.Context, meetingID, userID uuid.UUID, update *models.Attendance) error {
	return r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Updates(map[string]any{
			"is_present": update.IsPresent,
			"note":       update.Note,
			"updated_by": update.UpdatedBy,
		}).Error
}

// CountByBatchUser returns total meetings attended by a user in a batch
func (r *AttendanceRepository) CountByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Joins("JOIN meetings ON meetings.id = attendances.meeting_id").
		Where("attendances.user_id = ? AND attendances.is_present = TRUE AND meetings.batch_id = ?", userID, batchID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
