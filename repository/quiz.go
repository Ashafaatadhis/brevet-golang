package repository

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IQuizRepository interface
type IQuizRepository interface {
	WithTx(tx *gorm.DB) IQuizRepository
	GetQuizByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error)
	GetAllUpcomingQuizzes(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Quiz, int64, error)
	GetQuizByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Quiz, int64, error)
	GetQuestionByID(ctx context.Context, questionID, quizID uuid.UUID) (*models.QuizQuestion, error)
	Create(ctx context.Context, quiz *models.Quiz) error
	CreateOptions(ctx context.Context, options []models.QuizOption) error
	CreateQuestion(ctx context.Context, quiz *models.QuizQuestion) error
	SaveTempSubmission(ctx context.Context, temp *models.QuizTempSubmission) error
	CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error
	SaveQuizSubmission(ctx context.Context, sub *models.QuizSubmission) error
	UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error
	GetQuizWithQuestions(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error)
	CountQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) (int64, error)
	GetActiveAttemptByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*models.QuizAttempt, error)
	GetQuizAttemptByID(ctx context.Context, attemptID uuid.UUID) (*models.QuizAttempt, error)
	CreateQuizResult(ctx context.Context, result *models.QuizResult) error
	GetTempSubmissionsByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]models.QuizTempSubmission, error)
	GetAttemptByID(ctx context.Context, attemptID uuid.UUID) (*models.QuizAttempt, error)
	GetOptionByID(ctx context.Context, optionID, questionID uuid.UUID) (*models.QuizOption, error)
	GetAttemptsByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) ([]models.QuizAttempt, error)
	UpdateQuiz(ctx context.Context, quiz *models.Quiz) error
	DeleteQuiz(ctx context.Context, quizID uuid.UUID) error
	GetQuizResultByAttemptID(ctx context.Context, attemptID uuid.UUID) (*models.QuizResult, error)
	CountByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error)
	CountCompletedByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error)
	GetQuizzesWithScoresByBatchUser(ctx context.Context, batchID, userID uuid.UUID) ([]dto.QuizScore, error)
	GetAllByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]models.Quiz, error)
	GetQuizSubmissionByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*models.QuizSubmission, error)
}

// QuizRepository is a struct that represents a quiz repository
type QuizRepository struct {
	db *gorm.DB
}

// NewQuizRepository creates a new quiz repository
func NewQuizRepository(db *gorm.DB) IQuizRepository {
	return &QuizRepository{db: db}
}

// WithTx running with transaction
func (r *QuizRepository) WithTx(tx *gorm.DB) IQuizRepository {
	return &QuizRepository{db: tx}
}

// GetQuizByMeetingIDFiltered retrieves all quiz with pagination and filtering options
func (r *QuizRepository) GetQuizByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions) ([]models.Quiz, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Quiz{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Quiz{}).Where("meeting_id = ?", meetingID)

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "quizzes", opts.Filters, validSortFields, joinConditions, joinedRelations)

	var total int64
	db.Count(&total)

	var quizzes []models.Quiz
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&quizzes).Error

	return quizzes, total, err
}

// GetAllByMeetingID ambil semua quiz di meeting tertentu
func (r *QuizRepository) GetAllByMeetingID(ctx context.Context, meetingID uuid.UUID) ([]models.Quiz, error) {
	var quizzes []models.Quiz
	if err := r.db.WithContext(ctx).
		Where("meeting_id = ?", meetingID).
		Find(&quizzes).Error; err != nil {
		return nil, err
	}
	return quizzes, nil
}

// GetByQuizAndUser ambil submission quiz berdasarkan quiz_id & user_id
func (r *QuizRepository) GetQuizSubmissionByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*models.QuizSubmission, error) {
	var sub models.QuizSubmission
	if err := r.db.WithContext(ctx).
		Where("quiz_id = ? AND user_id = ?", quizID, userID).
		First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetAllUpcomingQuizzes retrieves upcoming quizzes that the user hasn't attempted yet
func (r *QuizRepository) GetAllUpcomingQuizzes(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Quiz, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Quiz{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "quizzes.end_at"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// db := r.db.WithContext(ctx).
	// 	Model(&models.Quiz{}).

	// 	Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
	// 	Joins("JOIN batches ON batches.id = meetings.batch_id").
	// 	Joins("JOIN purchases ON purchases.batch_id = batches.id").
	// 	// relasi ke quiz_attempts
	// 	Joins("LEFT JOIN quiz_attempts ON quiz_attempts.quiz_id = quizzes.id AND quiz_attempts.user_id = ?", userID).
	// 	// relasi ke quiz_submissions dari attempt tsb
	// 	Joins("LEFT JOIN quiz_submissions ON quiz_submissions.attempt_id = quiz_attempts.id").
	// 	Where("purchases.user_id = ? AND purchases.payment_status = ?", userID, models.Paid).
	// 	// belum pernah attempt, atau sudah attempt tapi belum submit
	// 	Where("(quiz_attempts.id IS NULL OR quiz_submissions.id IS NULL)").
	// 	// quiz masih aktif
	// 	Where("quizzes.end_time > ?", time.Now())

	db := r.db.WithContext(ctx).
		Model(&models.Quiz{}).
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Joins("JOIN batches ON batches.id = meetings.batch_id").
		Joins("JOIN purchases ON purchases.batch_id = batches.id").
		// cek attempt
		Joins("LEFT JOIN quiz_attempts ON quiz_attempts.quiz_id = quizzes.id AND quiz_attempts.user_id = ?", userID).
		Where("purchases.user_id = ? AND purchases.payment_status = ?", userID, models.Paid).
		// belum pernah attempt atau attempt belum selesai
		Where("quiz_attempts.id IS NULL OR quiz_attempts.ended_at IS NULL").
		// quiz masih aktif
		Where("quizzes.end_time > ?", time.Now())

	// Apply filters
	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}
	db = utils.ApplyFiltersWithJoins(db, "quizzes", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// Search by quiz title or meeting title
	if opts.Search != "" {
		db = db.Where("quizzes.title ILIKE ? OR meetings.title ILIKE ?", "%"+opts.Search+"%", "%"+opts.Search+"%")
	}

	var total int64
	db.Count(&total)

	var quizzes []models.Quiz
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Find(&quizzes).Error

	return quizzes, total, err
}

// GetQuizzesWithScoresByBatchUser retrieves quiz scores of a user in a batch
func (r *QuizRepository) GetQuizzesWithScoresByBatchUser(ctx context.Context, batchID, userID uuid.UUID) ([]dto.QuizScore, error) {
	var results []dto.QuizScore

	err := r.db.WithContext(ctx).
		Table("quizzes").
		Select(`
			quizzes.*,
			COALESCE(MAX(qr.score_percent), 0) AS score
		`).
		Joins("JOIN meetings m ON m.id = quizzes.meeting_id").
		Joins("LEFT JOIN quiz_attempts qa ON qa.quiz_id = quizzes.id AND qa.user_id = ?", userID).
		Joins("LEFT JOIN quiz_results qr ON qr.attempt_id = qa.id").
		Where("m.batch_id = ?", batchID).
		Group("quizzes.id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// CountQuestionsByQuizID menghitung jumlah soal di quiz tertentu
func (r *QuizRepository) CountQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.QuizQuestion{}).
		Where("quiz_id = ?", quizID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetQuestionByID retrieves a queestion by its ID
func (r *QuizRepository) GetQuestionByID(ctx context.Context, questionID, quizID uuid.UUID) (*models.QuizQuestion, error) {
	var q models.QuizQuestion
	err := r.db.WithContext(ctx).Where("id = ? AND quiz_id = ?", questionID, quizID).First(&q).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("question not found in this quiz")
		}
		return nil, err
	}
	return &q, nil
}

// GetTempSubmissionsByAttemptID ambil semua temp submission berdasarkan attemptID
func (r *QuizRepository) GetTempSubmissionsByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]models.QuizTempSubmission, error) {
	var temps []models.QuizTempSubmission

	err := r.db.WithContext(ctx).
		Preload("Question").
		Preload("Question.Options").
		Where("attempt_id = ?", attemptID).
		Find(&temps).Error

	if err != nil {
		return nil, err
	}
	return temps, nil
}

// GetQuizByID retrieves a quiz by its ID
func (r *QuizRepository) GetQuizByID(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error) {
	var quiz models.Quiz
	if err := r.db.WithContext(ctx).First(&quiz, "id = ?", quizID).Error; err != nil {
		return nil, err
	}
	return &quiz, nil
}

// Create for creating a new quiz
func (r *QuizRepository) Create(ctx context.Context, quiz *models.Quiz) error {
	return r.db.WithContext(ctx).Create(quiz).Error
}

// CreateQuestion for creating a new quiz question
func (r *QuizRepository) CreateQuestion(ctx context.Context, quiz *models.QuizQuestion) error {
	return r.db.WithContext(ctx).Create(quiz).Error
}

// CreateOptions untuk membuat banyak quiz option sekaligus
func (r *QuizRepository) CreateOptions(ctx context.Context, options []models.QuizOption) error {
	if len(options) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&options).Error
}

// CreateQuizAttempt create a quiz attempt
func (r *QuizRepository) CreateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

// SaveTempSubmission saves a temporary submission
func (r *QuizRepository) SaveTempSubmission(ctx context.Context, temp *models.QuizTempSubmission) error {
	var existing models.QuizTempSubmission

	// Cek apakah temp submission sudah ada untuk attempt dan question ini
	err := r.db.WithContext(ctx).
		Where("attempt_id = ? AND question_id = ?", temp.AttemptID, temp.QuestionID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Belum ada, buat baru
			return r.db.WithContext(ctx).Create(temp).Error
		}
		return err
	}

	// Sudah ada, update selected option
	existing.SelectedOptionID = temp.SelectedOptionID
	return r.db.WithContext(ctx).Save(&existing).Error
}

// GetActiveAttemptByQuizAndUser get attempt active
func (r *QuizRepository) GetActiveAttemptByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*models.QuizAttempt, error) {
	var attempt models.QuizAttempt
	err := r.db.WithContext(ctx).
		Where("quiz_id = ? AND user_id = ? AND ended_at IS NULL", quizID, userID).
		Order("started_at DESC").Limit(1).
		First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

// GetAttemptsByQuizAndUser ambil semua attempt user untuk quiz tertentu
func (r *QuizRepository) GetAttemptsByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) ([]models.QuizAttempt, error) {
	var attempts []models.QuizAttempt
	if err := r.db.WithContext(ctx).Where("quiz_id = ? AND user_id = ?", quizID, userID).Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

// SaveQuizSubmission save
func (r *QuizRepository) SaveQuizSubmission(ctx context.Context, sub *models.QuizSubmission) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// UpdateQuizAttempt update
func (r *QuizRepository) UpdateQuizAttempt(ctx context.Context, attempt *models.QuizAttempt) error {
	return r.db.WithContext(ctx).Save(attempt).Error
}

// GetQuizWithQuestions get quiz with questions
func (r *QuizRepository) GetQuizWithQuestions(ctx context.Context, quizID uuid.UUID) (*models.Quiz, error) {
	var quiz models.Quiz
	if err := r.db.WithContext(ctx).
		Preload("Questions.Options").
		First(&quiz, "id = ?", quizID).Error; err != nil {
		return nil, err
	}
	return &quiz, nil
}

// GetQuizAttemptByID get attempt by id
func (r *QuizRepository) GetQuizAttemptByID(ctx context.Context, attemptID uuid.UUID) (*models.QuizAttempt, error) {
	var attempt models.QuizAttempt
	if err := r.db.WithContext(ctx).
		Preload("Quiz").
		Where("id = ?", attemptID).
		First(&attempt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("quiz attempt not found")
		}
		return nil, err
	}
	return &attempt, nil
}

// CreateQuizResult create quiz result
func (r *QuizRepository) CreateQuizResult(ctx context.Context, result *models.QuizResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

// GetAttemptByID by id
func (r *QuizRepository) GetAttemptByID(ctx context.Context, attemptID uuid.UUID) (*models.QuizAttempt, error) {
	var attempt models.QuizAttempt
	if err := r.db.WithContext(ctx).
		First(&attempt, "id = ?", attemptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	return &attempt, nil
}

// GetOptionByID fetches a QuizOption by ID and ensures it belongs to the given question
func (r *QuizRepository) GetOptionByID(ctx context.Context, optionID, questionID uuid.UUID) (*models.QuizOption, error) {
	var option models.QuizOption
	if err := r.db.WithContext(ctx).
		Where("id = ? AND question_id = ?", optionID, questionID).
		First(&option).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("option not found for this question")
		}
		return nil, err
	}
	return &option, nil
}

// UpdateQuiz update
func (r *QuizRepository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) error {
	return r.db.WithContext(ctx).Save(quiz).Error
}

// DeleteQuiz delete
func (r *QuizRepository) DeleteQuiz(ctx context.Context, quizID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", quizID).Delete(&models.Quiz{}).Error
}

// GetQuizResultByAttemptID get quiz result
func (r *QuizRepository) GetQuizResultByAttemptID(ctx context.Context, attemptID uuid.UUID) (*models.QuizResult, error) {
	var result models.QuizResult
	if err := r.db.WithContext(ctx).
		Preload("Attempt.Quiz").
		First(&result, "attempt_id = ?", attemptID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// CountByBatchID count quizzes by batch ID
func (r *QuizRepository) CountByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Quiz{}).
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Where("meetings.batch_id = ?", batchID).
		Count(&count).Error
	return count, err
}

// CountCompletedByBatchUser counts completed quizzes for a user in a specific batch
func (r *QuizRepository) CountCompletedByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.QuizAttempt{}).
		Joins("JOIN quizzes ON quizzes.id = quiz_attempts.quiz_id").
		Joins("JOIN meetings ON meetings.id = quizzes.meeting_id").
		Where("meetings.batch_id = ? AND quiz_attempts.user_id = ? AND quiz_attempts.ended_at IS NOT NULL", batchID, userID).
		Distinct("quiz_attempts.quiz_id").
		Count(&count).Error
	return count, err
}
