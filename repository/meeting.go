package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MeetingRepository is a struct that represents a meeting repository
type MeetingRepository struct {
	db *gorm.DB
}

// NewMeetingRepository creates a new meeting repository
func NewMeetingRepository(db *gorm.DB) *MeetingRepository {
	return &MeetingRepository{db: db}
}

// WithTx running with transaction
func (r *MeetingRepository) WithTx(tx *gorm.DB) *MeetingRepository {
	return &MeetingRepository{db: tx}
}

// GetAllFilteredMeetings retrieves all meetings with pagination and filtering options
func (r *MeetingRepository) GetAllFilteredMeetings(opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields, err := utils.GetValidColumns(r.db, &models.Meeting{})
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

	db := r.db.Model(&models.Meeting{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "meetings", opts.Filters, validSortFields, joinConditions, joinedRelations)

	if opts.Search != "" {
		db = db.Where("title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var meetings []models.Meeting
	err = db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&meetings).Error

	return meetings, total, err
}

// GetMeetingsByBatchSlugFiltered retrieves all meetings with pagination and filtering options
func (r *MeetingRepository) GetMeetingsByBatchSlugFiltered(batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Meeting{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Model(&models.Meeting{}).
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

// FindByID retrieves a meeting by its ID
func (r *MeetingRepository) FindByID(id uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	err := r.db.First(&meeting, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// Create creates a new meeetings
func (r *MeetingRepository) Create(meeting *models.Meeting) error {
	return r.db.Create(meeting).Error
}

// Update updates an existing meeting
func (r *MeetingRepository) Update(meeting *models.Meeting) error {
	return r.db.Save(meeting).Error
}

// DeleteByID deletes a meeting by its ID
func (r *MeetingRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Meeting{}).Error
}

// AddTeachers is repo for add teacher to meeting
func (r *MeetingRepository) AddTeachers(meetingID uuid.UUID, teacherIDs []uuid.UUID) (*models.Meeting, error) {
	var teachers []models.User
	if err := r.db.Where("id IN ?", teacherIDs).Find(&teachers).Error; err != nil {
		return nil, err
	}

	var meeting models.Meeting
	if err := r.db.Preload("Teachers").Where("id = ?", meetingID).First(&meeting).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&meeting).Association("Teachers").Append(teachers); err != nil {
		return nil, err
	}

	// Refresh preload setelah Append
	if err := r.db.Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// GetTeacherIDsByMeetingID that repo function where's get teacher and pluck
func (r *MeetingRepository) GetTeacherIDsByMeetingID(meetingID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.
		Table("meeting_teachers").
		Where("meeting_id = ?", meetingID).
		Pluck("user_id", &ids).Error
	return ids, err
}

// UpdateTeachers this function repo to update teachers by meeting id and replae by array of teacher ids
func (r *MeetingRepository) UpdateTeachers(meetingID uuid.UUID, newTeacherIDs []uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := r.db.Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	var newTeachers []models.User
	if err := r.db.Where("id IN ?", newTeacherIDs).Find(&newTeachers).Error; err != nil {
		return nil, err
	}

	// Ganti semua guru dengan yang baru
	if err := r.db.Model(&meeting).Association("Teachers").Replace(newTeachers); err != nil {
		return nil, err
	}

	// Reload
	if err := r.db.Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// RemoveTeacher this repo function to remove teacher from meeting by meetingID
func (r *MeetingRepository) RemoveTeacher(meetingID uuid.UUID, teacherID uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := r.db.Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	var teachersToRemove models.User
	if err := r.db.Where("id = ?", teacherID).Find(&teachersToRemove).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&meeting).Association("Teachers").Delete(teachersToRemove); err != nil {
		return nil, err
	}

	// Reload
	if err := r.db.Preload("Teachers").First(&meeting, "id = ?", meetingID).Error; err != nil {
		return nil, err
	}

	return &meeting, nil
}

// GetTeachersByMeetingIDFiltered returns paginated + filtered list of teachers for a meeting
func (r *MeetingRepository) GetTeachersByMeetingIDFiltered(meetingID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
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

	db := r.db.
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
func (r *MeetingRepository) GetStudentsByBatchSlugFiltered(batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.User{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Preload("Profile").
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
func (r *MeetingRepository) IsBatchOwnedByUser(userID uuid.UUID, batchSlug string) (bool, error) {
	var count int64
	err := r.db.
		Model(&models.Meeting{}).
		Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("meeting_teachers.user_id = ? AND batches.slug = ?", userID, batchSlug).
		Count(&count).Error

	return count > 0, err
}

// GetMeetingsPurchasedByUserFiltered for get
func (r *MeetingRepository) GetMeetingsPurchasedByUserFiltered(userID uuid.UUID, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Meeting{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// JOIN ke purchases, dan preload relasi

	db := r.db.Preload("Teachers").
		Model(&models.Meeting{}).
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Joins("JOIN purchases ON purchases.batch_id = batches.id").
		Where("purchases.user_id = ?", userID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	// Apply dynamic filter (dari query param)
	db = utils.ApplyFiltersWithJoins(db, "batches", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// Search by title (opsional)
	if opts.Search != "" {
		db = db.Where("batches.title ILIKE ?", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var meeting []models.Meeting
	err := db.
		Order(fmt.Sprintf("meetings.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&meeting).Error

	return meeting, total, err
}

// GetMeetingsTaughtByTeacherFiltered for get
func (r *MeetingRepository) GetMeetingsTaughtByTeacherFiltered(userID uuid.UUID, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Meeting{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// 🔍 Subquery: cari batch.id di mana guru ngajar di salah satu meeting
	subQuery := r.db.
		Table("meetings").
		Select("batch_id").
		Joins("JOIN meeting_teachers ON meeting_teachers.meeting_id = meetings.id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("batches.slug = ? AND meeting_teachers.user_id = ?", batchSlug, userID).
		Group("batch_id")

	// 🔎 Ambil semua meetings di batch tersebut
	db := r.db.Preload("Teachers").
		Model(&models.Meeting{}).
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Where("meetings.batch_id IN (?)", subQuery).
		Where("batches.slug = ?", batchSlug)

	// 🔧 Dynamic filters
	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}
	db = utils.ApplyFiltersWithJoins(db, "meetings", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// 🔍 Search by meeting title
	if opts.Search != "" {
		q := "%" + opts.Search + "%"
		db = db.Where("meetings.title ILIKE ?", q)
	}

	// 🔢 Total count
	var total int64
	db.Count(&total)

	// 📦 Paginated result
	var meetings []models.Meeting
	err := db.
		Order(fmt.Sprintf("meetings.%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&meetings).Error

	return meetings, total, err
}

// IsMeetingTaughtByUser for get taught
func (r *MeetingRepository) IsMeetingTaughtByUser(meetingID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.
		Table("meeting_teachers").
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	return count > 0, err
}
