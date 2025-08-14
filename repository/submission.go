package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ISubmisssionRepository interface
type ISubmisssionRepository interface {
	WithTx(tx *gorm.DB) ISubmisssionRepository
	GetAllByAssignment(ctx context.Context, assignmentID uuid.UUID, userID *uuid.UUID, opts utils.QueryOptions) ([]models.AssignmentSubmission, int64, error)
	GetByIDAssignmentUser(ctx context.Context, submissionID, assignmentID, userID uuid.UUID) (models.AssignmentSubmission, error)
	Create(ctx context.Context, submission *models.AssignmentSubmission) error
	CreateSubmissionFiles(ctx context.Context, files []models.SubmissionFile) error
	GetByAssignmentUser(ctx context.Context, assignmentID, userID uuid.UUID) (models.AssignmentSubmission, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.AssignmentSubmission, error)
	Update(ctx context.Context, submission *models.AssignmentSubmission) error
	DeleteFilesBySubmissionID(ctx context.Context, submissionID uuid.UUID) error
	CreateFiles(ctx context.Context, files []models.SubmissionFile) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
	GetByIDUser(ctx context.Context, submissionID, userID uuid.UUID) (*models.AssignmentSubmission, error)
	GetGradeBySubmissionID(ctx context.Context, submissionID uuid.UUID) (*models.AssignmentGrade, error)
	UpsertGrade(ctx context.Context, grade models.AssignmentGrade) (models.AssignmentGrade, error)
	CountCompletedByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error)
}

// SubmissionRepository provides methods for managing submissions
type SubmissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates a new submissions repository
func NewSubmissionRepository(db *gorm.DB) ISubmisssionRepository {
	return &SubmissionRepository{db: db}
}

// WithTx running with transaction
func (r *SubmissionRepository) WithTx(tx *gorm.DB) ISubmisssionRepository {
	return &SubmissionRepository{db: tx}
}

// GetAllByAssignment for get all guru
func (r *SubmissionRepository) GetAllByAssignment(ctx context.Context, assignmentID uuid.UUID, userID *uuid.UUID, opts utils.QueryOptions) ([]models.AssignmentSubmission, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.AssignmentSubmission{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "created_at"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	db := r.db.WithContext(ctx).Preload("SubmissionFiles").
		Where("assignment_id = ?", assignmentID).
		Model(&models.AssignmentSubmission{})

	// Filter user_id kalau dikasih
	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}

	db = utils.ApplyFiltersWithJoins(db, "assignment_submissions", opts.Filters, validSortFields, map[string]string{}, map[string]bool{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var submissions []models.AssignmentSubmission
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&submissions).Error

	return submissions, total, err
}

// GetByIDAssignmentUser for get
func (r *SubmissionRepository) GetByIDAssignmentUser(ctx context.Context, submissionID, assignmentID, userID uuid.UUID) (models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := r.db.WithContext(ctx).Preload("SubmissionFiles").
		Where("id = ? AND assignment_id = ? AND user_id = ?", submissionID, assignmentID, userID).
		First(&submission).Error

	return submission, err
}

// Create is for create assignment_submissions
func (r *SubmissionRepository) Create(ctx context.Context, submission *models.AssignmentSubmission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

// CreateSubmissionFiles for create submission_files
func (r *SubmissionRepository) CreateSubmissionFiles(ctx context.Context, files []models.SubmissionFile) error {
	return r.db.WithContext(ctx).Create(&files).Error
}

// GetByAssignmentUser is get submission by assignment id and user id
func (r *SubmissionRepository) GetByAssignmentUser(ctx context.Context, assignmentID, userID uuid.UUID) (models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := r.db.WithContext(ctx).Where("assignment_id = ? AND user_id = ?", assignmentID, userID).First(&submission).Error
	return submission, err
}

// FindByID get submission by id with preload submissionFiles
func (r *SubmissionRepository) FindByID(ctx context.Context, id uuid.UUID) (models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := r.db.WithContext(ctx).Preload("SubmissionFiles").Where("id = ?", id).First(&submission).Error
	return submission, err
}

// Update for update
func (r *SubmissionRepository) Update(ctx context.Context, submission *models.AssignmentSubmission) error {
	return r.db.WithContext(ctx).Save(submission).Error
}

// DeleteFilesBySubmissionID deletes file submission by assignment_submission_id
func (r *SubmissionRepository) DeleteFilesBySubmissionID(ctx context.Context, submissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("assignment_submission_id = ?", submissionID).Delete(&models.SubmissionFile{}).Error
}

// CreateFiles is create files
func (r *SubmissionRepository) CreateFiles(ctx context.Context, files []models.SubmissionFile) error {
	return r.db.WithContext(ctx).Create(&files).Error
}

// DeleteByID deletes a submission by its ID
func (r *SubmissionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.AssignmentSubmission{}).
		Error
}

// GetByIDUser get submission by id and user id
func (r *SubmissionRepository) GetByIDUser(ctx context.Context, submissionID, userID uuid.UUID) (*models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	err := r.db.WithContext(ctx).
		Preload("SubmissionFiles").
		Where("id = ? AND user_id = ?", submissionID, userID).
		First(&submission).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

// GetGradeBySubmissionID repo get grade by submission id
func (r *SubmissionRepository) GetGradeBySubmissionID(ctx context.Context, submissionID uuid.UUID) (*models.AssignmentGrade, error) {
	var grade models.AssignmentGrade
	err := r.db.WithContext(ctx).Preload("GradedByUser").
		Where("assignment_submission_id = ?", submissionID).
		First(&grade).Error

	if err != nil {
		return nil, err
	}

	return &grade, nil
}

// UpsertGrade for upsert
func (r *SubmissionRepository) UpsertGrade(ctx context.Context, grade models.AssignmentGrade) (models.AssignmentGrade, error) {
	var existing models.AssignmentGrade
	err := r.db.WithContext(ctx).Where("assignment_submission_id = ?", grade.AssignmentSubmissionID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert baru
		if err := r.db.WithContext(ctx).Create(&grade).Error; err != nil {
			return models.AssignmentGrade{}, err
		}
		return grade, nil
	} else if err != nil {
		return models.AssignmentGrade{}, err
	}

	// Update kalau sudah ada
	existing.Grade = grade.Grade
	existing.Feedback = grade.Feedback
	existing.GradedBy = grade.GradedBy

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return models.AssignmentGrade{}, err
	}
	return existing, nil
}

// CountCompletedByBatchUser for count completed submission by batch id and user id
func (r *SubmissionRepository) CountCompletedByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AssignmentSubmission{}).
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Joins("JOIN meetings ON meetings.id = assignments.meeting_id").
		Where("meetings.batch_id = ? AND assignment_submissions.user_id = ?", batchID, userID).
		Count(&count).Error
	return count, err
}
