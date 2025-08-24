package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// IQuizService interface
type IQuizService interface {
	GetQuizByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions, user *utils.Claims) ([]models.Quiz, int64, error)
	ImportQuestionsFromExcel(ctx context.Context, user *utils.Claims, quizID uuid.UUID, fileHeader *multipart.FileHeader) error
	CreateQuizMetadata(ctx context.Context, user *utils.Claims, meetingID uuid.UUID, req *dto.ImportQuizzesRequest) (*models.Quiz, error)
	SaveTempSubmission(ctx context.Context, user *utils.Claims, attemptID uuid.UUID, body *dto.SaveTempSubmissionRequest) error
	StartQuiz(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.QuizAttempt, error)
	SubmitQuiz(ctx context.Context, user *utils.Claims, attemptID uuid.UUID) error
	GetQuizMetadata(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.Quiz, error)
	GetQuizWithQuestions(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.Quiz, error)
	GetActiveAttempt(ctx context.Context, quizID uuid.UUID, user *utils.Claims) (*models.QuizAttempt, error)
	GetAttemptDetail(ctx context.Context, attemptID uuid.UUID, user *utils.Claims) (*dto.QuizAttemptFull, error)
	UpdateQuiz(ctx context.Context, quizID uuid.UUID, user *utils.Claims, body *dto.UpdateQuizRequest) (*models.Quiz, error)
	DeleteQuiz(ctx context.Context, quizID uuid.UUID, user *utils.Claims) error
	GetAttemptResult(ctx context.Context, attemptID uuid.UUID, user *utils.Claims) (*models.QuizResult, error)
}

// QuizService provides methods for managing quizzes
type QuizService struct {
	quizRepo        repository.IQuizRepository
	batchRepo       repository.IBatchRepository
	meetingRepo     repository.IMeetingRepository
	purchaseService IPurchaseService
	fileService     IFileService
	db              *gorm.DB
}

// NewQuizService creates a new instance of QuizService
func NewQuizService(quizRepo repository.IQuizRepository, batchRepo repository.IBatchRepository, meetingRepo repository.IMeetingRepository, purchaseService IPurchaseService, fileService IFileService, db *gorm.DB) IQuizService {
	return &QuizService{quizRepo: quizRepo, batchRepo: batchRepo, meetingRepo: meetingRepo, purchaseService: purchaseService, fileService: fileService, db: db}
}

func (s *QuizService) checkUserAccess(ctx context.Context, user *utils.Claims, meetingID uuid.UUID) (bool, error) {
	// Cari batch info dari meetingID
	batch, err := s.batchRepo.GetBatchByMeetingID(ctx, meetingID) // balikin batchSlug & batchID
	if err != nil {
		return false, err
	}

	// Kalau role teacher, cek apakah dia mengajar batch ini
	if user.Role == string(models.RoleTypeGuru) {
		return s.meetingRepo.IsBatchOwnedByUser(ctx, user.UserID, batch.Slug)
	}

	// Kalau student, cek pembayaran
	if user.Role == string(models.RoleTypeSiswa) {
		return s.purchaseService.HasPaid(ctx, user.UserID, batch.ID)
	}

	// Role lain tidak diizinkan
	return false, nil
}

// StartQuiz start quiz
func (s *QuizService) StartQuiz(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.QuizAttempt, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// cek akses user ke quiz
	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)

	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// cek apakah quiz sedang dibuka
	now := time.Now()
	if !quiz.IsOpen {
		return nil, fmt.Errorf("quiz is not open")
	}
	if now.Before(quiz.StartTime) {
		return nil, fmt.Errorf("quiz has not started yet")
	}
	if !quiz.EndTime.IsZero() && now.After(quiz.EndTime) {
		return nil, fmt.Errorf("quiz has ended")
	}

	// ambil semua attempt user untuk quiz ini
	attempts, err := s.quizRepo.GetAttemptsByQuizAndUser(ctx, quizID, user.UserID)
	if err != nil {
		return nil, err
	}

	// cek jumlah attempt
	if quiz.MaxAttempts > 0 && len(attempts) >= quiz.MaxAttempts {
		return nil, fmt.Errorf("maximum attempts reached")
	}

	// cek apakah ada attempt aktif
	activeAttempt, err := s.quizRepo.GetActiveAttemptByQuizAndUser(ctx, quizID, user.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if activeAttempt != nil {
		return nil, fmt.Errorf("quiz attempt still ongoing")
	}

	// buat attempt baru
	attempt := &models.QuizAttempt{
		UserID:    user.UserID,
		QuizID:    quizID,
		StartedAt: now,
	}

	if err := s.quizRepo.CreateQuizAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	// ambil ulang attempt lengkap
	attempt, err = s.quizRepo.GetAttemptByID(ctx, attempt.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("attempt not found")
		}
		return nil, err
	}

	return attempt, nil
}

// ImportQuestionsFromExcel excel
func (s *QuizService) ImportQuestionsFromExcel(
	ctx context.Context,
	user *utils.Claims,
	quizID uuid.UUID,
	fileHeader *multipart.FileHeader,
) error {
	// cek quiz exists + akses
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return err
	}

	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden: not teacher of this meeting")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return err
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}
	if len(rows) <= 1 {
		return fmt.Errorf("excel kosong / tidak ada soal")
	}

	// transaksi DB
	return utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		for i := 1; i < len(rows); i++ {
			if len(rows[i]) < 3 {
				continue
			}

			questionText := rows[i][0]
			correctLetter := strings.ToUpper(rows[i][len(rows[i])-1])
			optionCols := rows[i][1 : len(rows[i])-1]

			// simpan ke tabel quiz_questions
			q := models.QuizQuestion{
				ID:       uuid.New(),
				QuizID:   quiz.ID,
				Question: questionText,
			}
			if err := s.quizRepo.WithTx(tx).CreateQuestion(ctx, &q); err != nil {
				return fmt.Errorf("row %d: %w", i+1, err)
			}

			// buat opsi
			var options []models.QuizOption
			for idx, optText := range optionCols {
				letter := string(rune('A' + idx))
				options = append(options, models.QuizOption{
					ID:         uuid.New(),
					QuestionID: q.ID,
					OptionText: optText,
					IsCorrect:  (letter == correctLetter),
				})
			}

			if err := s.quizRepo.WithTx(tx).CreateOptions(ctx, options); err != nil {
				return fmt.Errorf("row %d: %w", i+1, err)
			}
		}
		return nil
	})
}

// CreateQuizMetadata for create
func (s *QuizService) CreateQuizMetadata(
	ctx context.Context,
	user *utils.Claims,
	meetingID uuid.UUID,
	req *dto.ImportQuizzesRequest,
) (*models.Quiz, error) {

	// cek akses
	allowed, err := s.checkUserAccess(ctx, user, meetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: not teacher of this meeting")
	}

	quiz := &models.Quiz{
		ID:             uuid.New(),
		MeetingID:      meetingID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		IsOpen:         false,
		MaxAttempts:    req.MaxAttempts,
		DurationMinute: req.DurationMinute,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
	}

	if err := s.quizRepo.Create(ctx, quiz); err != nil {
		return nil, err
	}

	return quiz, nil
}

// SubmitQuiz submit quiz
func (s *QuizService) SubmitQuiz(ctx context.Context, user *utils.Claims, attemptID uuid.UUID) error {
	return utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		// 1️⃣ Ambil attempt berdasarkan attemptID
		attempt, err := s.quizRepo.WithTx(tx).GetQuizAttemptByID(ctx, attemptID)
		if err != nil {
			return err
		}

		// 2️⃣ Pastikan attempt milik user
		if attempt.UserID != user.UserID {
			return fmt.Errorf("forbidden: not your attempt")
		}

		// 3️⃣ Pastikan attempt belum selesai
		if attempt.EndedAt != nil {
			return fmt.Errorf("quiz already submitted")
		}

		quiz, err := s.quizRepo.WithTx(tx).GetQuizByID(ctx, attempt.QuizID)
		if err != nil {
			return err
		}

		allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("forbidden")
		}

		// 4️⃣ Ambil semua temp submission milik attempt ini
		temps, err := s.quizRepo.WithTx(tx).GetTempSubmissionsByAttemptID(ctx, attemptID)
		if err != nil {
			return err
		}

		if len(temps) == 0 {
			return fmt.Errorf("no submissions found")
		}

		// 5️⃣ Simpan final submission & hitung total benar
		correctAnswers := 0
		for _, temp := range temps {
			score := 0
			for _, opt := range temp.Question.Options {
				if opt.ID == temp.SelectedOptionID && opt.IsCorrect {
					score = 1
					correctAnswers++
					break
				}
			}

			sub := &models.QuizSubmission{
				AttemptID:        attempt.ID,
				QuestionID:       temp.QuestionID,
				SelectedOptionID: temp.SelectedOptionID,
				Score:            score,
			}

			if err := s.quizRepo.WithTx(tx).SaveQuizSubmission(ctx, sub); err != nil {
				return err
			}
		}

		totalQuestions := len(temps)
		wrongAnswers := totalQuestions - correctAnswers
		scorePercentage := (correctAnswers * 100) / totalQuestions

		// 6️⃣ Simpan QuizResult
		quizResult := &models.QuizResult{
			ID:             uuid.New(),
			AttemptID:      attempt.ID,
			TotalQuestions: totalQuestions,
			CorrectAnswers: correctAnswers,
			WrongAnswers:   wrongAnswers,
			ScorePercent:   float64(scorePercentage),
		}

		if err := s.quizRepo.WithTx(tx).CreateQuizResult(ctx, quizResult); err != nil {
			return err
		}

		// 7️⃣ Update attempt.EndedAt
		now := time.Now()
		attempt.EndedAt = &now
		if err := s.quizRepo.WithTx(tx).UpdateQuizAttempt(ctx, attempt); err != nil {
			return err
		}

		return nil
	})
}

// SaveTempSubmission service
// SaveTempSubmission service
func (s *QuizService) SaveTempSubmission(ctx context.Context, user *utils.Claims, attemptID uuid.UUID, body *dto.SaveTempSubmissionRequest) error {
	// 1️⃣ Ambil attempt
	attempt, err := s.quizRepo.GetQuizAttemptByID(ctx, attemptID)
	if err != nil {
		return err
	}

	// pastikan attempt punya user yang sama
	if attempt.UserID != user.UserID {
		return fmt.Errorf("forbidden: not your attempt")
	}

	// pastikan belum selesai
	if attempt.EndedAt != nil {
		return fmt.Errorf("quiz already submitted")
	}

	// 2️⃣ Ambil quiz dari attempt
	quiz, err := s.quizRepo.GetQuizByID(ctx, attempt.QuizID)
	if err != nil {
		return err
	}

	// 3️⃣ Cek akses ke quiz
	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden: not allowed to access this quiz")
	}

	// 4️⃣ Cek apakah question memang bagian dari quiz
	question, err := s.quizRepo.GetQuestionByID(ctx, body.QuestionID, quiz.ID)
	if err != nil {
		return fmt.Errorf("question not found in this quiz")
	}

	// 5️⃣ Cek apakah SelectedOptionID memang milik question tersebut
	option, err := s.quizRepo.GetOptionByID(ctx, body.SelectedOptionID, question.ID)
	if err != nil {
		return fmt.Errorf("selected option does not belong to the question")
	}

	// 6️⃣ Save atau update temp submission
	temp := &models.QuizTempSubmission{
		AttemptID:        attempt.ID,
		QuestionID:       question.ID,
		SelectedOptionID: option.ID,
	}

	return s.quizRepo.SaveTempSubmission(ctx, temp)
}

// GetQuizMetadata get quiz by id
func (s *QuizService) GetQuizMetadata(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// opsional: hide fields tertentu
	// quiz.Questions = nil
	return quiz, nil
}

// GetQuizByMeetingIDFiltered filtered
func (s *QuizService) GetQuizByMeetingIDFiltered(ctx context.Context, meetingID uuid.UUID, opts utils.QueryOptions, user *utils.Claims) ([]models.Quiz, int64, error) {

	allowed, err := s.checkUserAccess(ctx, user, meetingID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, fmt.Errorf("forbidden")
	}

	quizzes, total, err := s.quizRepo.GetQuizByMeetingIDFiltered(ctx, meetingID, opts)
	if err != nil {
		return nil, 0, err
	}

	return quizzes, total, nil

}

// GetQuizWithQuestions ambil quiz + questions + options + temp submission user
func (s *QuizService) GetQuizWithQuestions(ctx context.Context, user *utils.Claims, quizID uuid.UUID) (*models.Quiz, error) {

	// 2️⃣ Ambil quiz lengkap dari repo
	quiz, err := s.quizRepo.GetQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// 3️⃣ Opsional: cek apakah user boleh akses quiz ini
	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	return quiz, nil
}

// GetActiveAttempt get active attempt
func (s *QuizService) GetActiveAttempt(ctx context.Context, quizID uuid.UUID, user *utils.Claims) (*models.QuizAttempt, error) {
	quiz, err := s.quizRepo.GetQuizWithQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}
	// 3️⃣ Opsional: cek apakah user boleh akses quiz ini
	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	attempt, err := s.quizRepo.GetActiveAttemptByQuizAndUser(ctx, quizID, user.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return attempt, nil
}

// GetAttemptDetail detail
func (s *QuizService) GetAttemptDetail(ctx context.Context, attemptID uuid.UUID, user *utils.Claims) (*dto.QuizAttemptFull, error) {
	// Ambil attempt
	attempt, err := s.quizRepo.GetAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("attempt not found")
		}
		return nil, err
	}

	// Pastikan user punya akses ke attempt ini
	if attempt.UserID != user.UserID {
		return nil, fmt.Errorf("forbidden")
	}

	// Ambil quiz dan pertanyaan
	quiz, err := s.quizRepo.GetQuizWithQuestions(ctx, attempt.QuizID)
	if err != nil {
		return nil, err
	}

	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// Ambil temp submissions jika ada
	tempSubs, _ := s.quizRepo.GetTempSubmissionsByAttemptID(ctx, attempt.ID)

	return &dto.QuizAttemptFull{
		Attempt:         attempt,
		Quiz:            quiz,
		TempSubmissions: tempSubs,
	}, nil

}

// UpdateQuiz update
func (s *QuizService) UpdateQuiz(ctx context.Context, quizID uuid.UUID, user *utils.Claims, body *dto.UpdateQuizRequest) (*models.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// copy field dari DTO ke model (IgnoreEmpty supaya nil tidak menimpa)
	if err := copier.CopyWithOption(quiz, body, copier.Option{IgnoreEmpty: true, DeepCopy: true}); err != nil {
		return nil, fmt.Errorf("failed to copy quiz data: %w", err)
	}

	if err := s.quizRepo.UpdateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	return quiz, nil
}

// DeleteQuiz delete
func (s *QuizService) DeleteQuiz(ctx context.Context, quizID uuid.UUID, user *utils.Claims) error {
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return err
	}

	allowed, err := s.checkUserAccess(ctx, user, quiz.MeetingID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden")
	}

	return s.quizRepo.DeleteQuiz(ctx, quizID)
}

// GetAttemptResult result
func (s *QuizService) GetAttemptResult(ctx context.Context, attemptID uuid.UUID, user *utils.Claims) (*models.QuizResult, error) {
	result, err := s.quizRepo.GetQuizResultByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	fmt.Println("CEK MEETING ID : ", result.Attempt.Quiz.MeetingID)
	allowed, err := s.checkUserAccess(ctx, user, result.Attempt.Quiz.MeetingID)
	fmt.Println(allowed, err, "TEST TES")
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// cek user
	if user.Role == string(models.RoleTypeSiswa) && result.Attempt.UserID != user.UserID {
		return nil, fmt.Errorf("forbidden")
	}

	return result, nil
}
