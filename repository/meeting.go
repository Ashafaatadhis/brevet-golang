package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IMeetingRepository interface
type IMeetingRepository interface {
	WithTx(tx *gorm.DB) IMeetingRepository
	GetAllFilteredMeetings(ctx context.Context, opts utils.QueryOptions) ([]models.Meeting, int64, error)
	GetMeetingsByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error)
	GetMeetingsByBatchID(ctx context.Context, batchID uuid.UUID) ([]models.Meeting, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Meeting, error)
	Create(ctx context.Context, meeting *models.Meeting) error
	Update(ctx context.Context, meeting *models.Meeting) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	AddTeachers(ctx context.Context, meetingID uuid.UUID, teacherIDs []uuid.UUID) (*models.Meeting, error)
	GetTeacherIDsByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]uuid.UUID, error)
	UpdateTeachers(ctx context.Context, meetingID uuid.UUID, newTeacherIDs []uuid.UUID) (*models.Meeting, error)
	RemoveTeacher(ctx context.Context, meetingID uuid.UUID, teacherID uuid.UUID) (*models.Meeting, error)
	GetTeachersByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error)
	GetStudentsByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error)
	IsBatchOwnedByUser(ctx context.Context, userID uuid.UUID, batchSlug string) (bool, error)
	IsMeetingTaughtByUser(ctx context.Context, meetingID, userID uuid.UUID) (bool, error)
	IsUserTeachingInMeeting(ctx context.Context, userID, meetingID uuid.UUID) (bool, error)
	GetMeetingNamesByBatchID(ctx context.Context, batchID uuid.UUID) ([]string, error)
	GetPrevMeeting(ctx context.Context, batchID uuid.UUID, startAt time.Time) (*models.Meeting, error)
}

// MeetingRepository is a struct that represents a meeting repository
type MeetingRepository struct {
	db *gorm.DB
}

// NewMeetingRepository creates a new meeting repository
func NewMeetingRepository(db *gorm.DB) IMeetingRepository {
	return &MeetingRepository{db: db}
}

// WithTx running with transaction
func (r *MeetingRepository) WithTx(tx *gorm.DB) IMeetingRepository {
	return &MeetingRepository{db: tx}
}

// GetAllFilteredMeetings retrieves all meetings with pagination and filtering options
func (r *MeetingRepository) GetAllFilteredMeetings(ctx context.Context, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Meeting{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("Teachers").Preload("Quizzes").Preload("Materials").Preload("Assignments").Preload("Assignments.AssignmentFiles", func(db *gorm.DB) *gorm.DB {
		return db.Order("assignment_files.created_at ASC") // urut berdasarkan waktu upload paling awal
	}).
		Model(&models.Meeting{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "meetings", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var meetings []models.Meeting
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&meetings).Error

	return meetings, total, err
}

// GetMeetingsByBatchSlugFiltered retrieves all meetings with pagination and filtering options
func (r *MeetingRepository) GetMeetingsByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Meeting{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("Teachers").Preload("Quizzes").Preload("Materials").Preload("Assignments").Preload("Assignments.AssignmentFiles", func(db *gorm.DB) *gorm.DB {
		return db.Order("assignment_files.created_at ASC")
	}).
		Model(&models.Meeting{}).
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("batches.slug = ?", batchSlug)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "meetings", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var meetings []models.Meeting
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&meetings).Error

	return meetings, total, err
}

// GetMeetingsByBatchID retrieves all meetings belonging to a specific batch ID
func (r *MeetingRepository) GetMeetingsByBatchID(ctx context.Context, batchID uuid.UUID) ([]models.Meeting, error) {
	var meetings []models.Meeting
	err := r.db.WithContext(ctx).
		Where("batch_id = ?", batchID).
		Find(&meetings).Error
	if err != nil {
		return nil, err
	}
	return meetings, nil
}

func (r *MeetingRepository) GetPrevMeeting(ctx context.Context, batchID uuid.UUID, startAt time.Time) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := r.db.WithContext(ctx).
		Where("batch_id = ? AND start_at < ?", batchID, startAt).
		Order("start_at DESC").
		First(&meeting).Error; err != nil {
		return nil, err
	}
	return &meeting, nil
}

// FindByID retrieves a meeting by its ID
func (r *MeetingRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	err := r.db.WithContext(ctx).Preload("Teachers").Preload("Quizzes").Preload("Materials").Preload("Assignments").Preload("Assignments.AssignmentFiles", func(db *gorm.DB) *gorm.DB {
		return db.Order("assignment_files.created_at ASC") // urut berdasarkan waktu upload paling awal
	}).
		First(&meeting, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// GetMeetingNamesByBatchID mengambil daftar nama meeting berdasarkan BatchID
func (r *MeetingRepository) GetMeetingNamesByBatchID(ctx context.Context, batchID uuid.UUID) ([]string, error) {
	var names []string

	err := r.db.WithContext(ctx).
		Model(&models.Meeting{}).
		Select("meetings.title").
		Where("meetings.batch_id = ?", batchID).
		Order("meetings.created_at ASC").
		Pluck("meetings.title", &names).Error

	if err != nil {
		return nil, err
	}

	return names, nil
}

// Create creates a new meeetings
func (r *MeetingRepository) Create(ctx context.Context, meeting *models.Meeting) error {
	return r.db.WithContext(ctx).Create(meeting).Error
}

// Update updates an existing meeting
func (r *MeetingRepository) Update(ctx context.Context, meeting *models.Meeting) error {
	return r.db.WithContext(ctx).Save(meeting).Error
}

// DeleteByID deletes a meeting by its ID
func (r *MeetingRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Meeting{}).Error
}

// AddTeachers is repo for add teacher to meeting
func (r *MeetingRepository) AddTeachers(ctx context.Context, meetingID uuid.UUID, teacherIDs []uuid.UUID) (*models.Meeting, error) {
	var teachers []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", teacherIDs).Find(&teachers).Error; err != nil {
		return nil, err
	}

	var meeting models.Meeting
	if err := r.db.WithContext(ctx).Preload("Teachers").Where("id = ?", meetingID).First(&meeting).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&meeting).Association("Teachers").Append(teachers); err != nil {
		return nil, err
	}

	// Refresh preload setelah Append
	if err := r.db.WithContext(ctx).Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// GetTeacherIDsByMeetingID that repo function where's get teacher and pluck
func (r *MeetingRepository) GetTeacherIDsByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Table("meeting_teachers").
		Where("meeting_id = ?", meetingID).
		Pluck("user_id", &ids).Error
	return ids, err
}

// UpdateTeachers this function repo to update teachers by meeting id and replae by array of teacher ids
func (r *MeetingRepository) UpdateTeachers(ctx context.Context, meetingID uuid.UUID, newTeacherIDs []uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := r.db.WithContext(ctx).Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	var newTeachers []models.User
	if err := r.db.WithContext(ctx).Where("id IN ?", newTeacherIDs).Find(&newTeachers).Error; err != nil {
		return nil, err
	}

	// Ganti semua guru dengan yang baru
	if err := r.db.WithContext(ctx).Model(&meeting).Association("Teachers").Replace(newTeachers); err != nil {
		return nil, err
	}

	// Reload
	if err := r.db.WithContext(ctx).Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// IsUserTeachingInMeeting for know user is teacher in this meet
func (r *MeetingRepository) IsUserTeachingInMeeting(ctx context.Context, userID, meetingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("meeting_teachers").
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RemoveTeacher this repo function to remove teacher from meeting by meetingID
func (r *MeetingRepository) RemoveTeacher(ctx context.Context, meetingID uuid.UUID, teacherID uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := r.db.WithContext(ctx).Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	var teachersToRemove models.User
	if err := r.db.WithContext(ctx).Where("id = ?", teacherID).Find(&teachersToRemove).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&meeting).Association("Teachers").Delete(teachersToRemove); err != nil {
		return nil, err
	}

	// Reload
	if err := r.db.WithContext(ctx).Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// GetTeachersByMeetingIDFiltered returns paginated + filtered list of teachers for a meeting
func (r *MeetingRepository) GetTeachersByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.User{}, &models.MeetingTeacher{}, &models.Meeting{})
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

	db := r.db.WithContext(ctx).
		Model(&models.User{}).
		Joins("JOIN meeting_teachers ON meeting_teachers.user_id = users.id").
		Where("meeting_teachers.meeting_id = ?", meetingID)

	joinConditions := map[string]string{} // Tambahkan kalau ada relasi lain
	joinedRelations := map[string]bool{}  // Tracking relasi

	db = utils.ApplyFiltersWithJoins(db, "users", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		search := "%" + opts.Search + "%"
		db = db.Where("users.name ILIKE ? OR users.email ILIKE ?", search, search)
	}

	var total int64
	db.Count(&total)

	var teachers []models.User
	err = db.
		Order(fmt.Sprintf("users.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&teachers).Error

	return teachers, total, err
}

// GetStudentsByBatchSlugFiltered get all students by batch
func (r *MeetingRepository) GetStudentsByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.User{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Preload("Profile").
		Model(&models.User{}).
		Joins("JOIN purchases ON purchases.user_id = users.id").
		Joins("JOIN batches ON batches.id = purchases.batch_id").
		Where("batches.slug = ?", batchSlug).
		Where("users.role_type = ?", models.RoleTypeSiswa).
		Group("users.id")

	// Apply filters
	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}
	db = utils.ApplyFiltersWithJoins(db, "users", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// Search
	if opts.Search != "" {
		q := "%" + opts.Search + "%"
		db = db.Where("users.name ILIKE ? OR users.email ILIKE ?", q, q)
	}

	var total int64
	db.Count(&total)

	var students []models.User
	err := db.
		Order(fmt.Sprintf("users.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&students).Error

	return students, total, err
}

// IsBatchOwnedByUser for get all batch by owned teacher
func (r *MeetingRepository) IsBatchOwnedByUser(ctx context.Context, userID uuid.UUID, batchSlug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Meeting{}).
		Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("meeting_teachers.user_id = ? AND batches.slug = ?", userID, batchSlug).
		Count(&count).Error

	return count > 0, err
}

// IsMeetingTaughtByUser for get taught
func (r *MeetingRepository) IsMeetingTaughtByUser(ctx context.Context, meetingID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("meeting_teachers").
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	return count > 0, err
}
