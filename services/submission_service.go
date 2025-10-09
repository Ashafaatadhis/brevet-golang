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
	"strconv"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ISubmissionService interface
type ISubmissionService interface {
	checkUserAccess(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (bool, error)
	GetAllSubmissionsByAssignmentUser(ctx context.Context, assignmentID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.AssignmentSubmission, int64, error)
	GetSubmissionDetail(ctx context.Context, submissionID uuid.UUID, user *utils.Claims) (models.AssignmentSubmission, error)
	CreateSubmission(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, req *dto.CreateSubmissionRequest, fileURLs []string) (*models.AssignmentSubmission, error)
	UpdateSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID, body *dto.UpdateSubmissionRequest) (*models.AssignmentSubmission, error)
	DeleteSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID) error
	GetSubmissionGrade(ctx context.Context, user *utils.Claims, submissionID uuid.UUID) (*models.AssignmentGrade, error)
	GradeSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID, req *dto.GradeSubmissionRequest) (models.AssignmentGrade, error)
	GenerateGradesExcel(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (*excelize.File, string, error)
	ImportGradesFromExcel(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, fileHeader *multipart.FileHeader) error
}

// SubmissionService provides methods for managing submissions
type SubmissionService struct {
	submissionRepo  repository.ISubmisssionRepository
	assignmentRepo  repository.IAssignmentRepository
	meetingRepo     repository.IMeetingRepository
	attendanceRepo  repository.IAttendanceRepository
	quizRepo        repository.IQuizRepository
	purchaseService IPurchaseService
	fileService     IFileService
	db              *gorm.DB
}

// NewSubmissionService creates a new instance of SubmissionService
func NewSubmissionService(submissionRepo repository.ISubmisssionRepository, assignmentRepo repository.IAssignmentRepository,
	meetingRepo repository.IMeetingRepository, attendanceRepo repository.IAttendanceRepository,
	quizRepo repository.IQuizRepository, purchaseService IPurchaseService,
	fileService IFileService, db *gorm.DB) ISubmissionService {
	return &SubmissionService{submissionRepo: submissionRepo, assignmentRepo: assignmentRepo, attendanceRepo: attendanceRepo, quizRepo: quizRepo, meetingRepo: meetingRepo, purchaseService: purchaseService, fileService: fileService, db: db}
}

func (s *SubmissionService) checkUserAccess(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (bool, error) {
	// Cari batch info dari assignmentID
	batch, err := s.assignmentRepo.GetBatchByAssignmentID(ctx, assignmentID) // balikin batchSlug & batchID
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

	if user.Role == string(models.RoleTypeAdmin) {
		return true, nil
	}

	// Role lain tidak diizinkan
	return false, nil
}

// GetAllSubmissionsByAssignmentUser for get all
func (s *SubmissionService) GetAllSubmissionsByAssignmentUser(ctx context.Context, assignmentID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.AssignmentSubmission, int64, error) {
	allowed, err := s.checkUserAccess(ctx, user, assignmentID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, errors.New("user not authorized to access this assignment")
	}
	if user.Role == string(models.RoleTypeGuru) || user.Role == string(models.RoleTypeAdmin) {
		return s.submissionRepo.GetAllByAssignment(ctx, assignmentID, nil, opts)
	}
	return s.submissionRepo.GetAllByAssignment(ctx, assignmentID, &user.UserID, opts)

}

// GetSubmissionDetail fot get detail
func (s *SubmissionService) GetSubmissionDetail(ctx context.Context, submissionID uuid.UUID, user *utils.Claims) (models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	var err error

	if user.Role == string(models.RoleTypeGuru) {
		submission, err = s.submissionRepo.FindByID(ctx, submissionID)
	} else {
		subPtr, err2 := s.submissionRepo.GetByIDUser(ctx, submissionID, user.UserID)
		if err2 != nil {
			return models.AssignmentSubmission{}, err2
		}
		if subPtr != nil {
			submission = *subPtr
		}
		err = err2
	}
	if err != nil {
		return models.AssignmentSubmission{}, err
	}

	// Cek akses
	allowed, err := s.checkUserAccess(ctx, user, submission.AssignmentID)
	if err != nil {
		return models.AssignmentSubmission{}, err
	}
	if !allowed {
		return models.AssignmentSubmission{}, errors.New("user not authorized to access this assignment")
	}

	return submission, nil
}

// CreateSubmission is for create submission
// func (s *SubmissionService) CreateSubmission(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, req *dto.CreateSubmissionRequest, fileURLs []string) (*models.AssignmentSubmission, error) {
// 	var submission models.AssignmentSubmission

// 	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
// 		// Cek apakah user sudah bayar batch terkait assignment
// 		allowed, err := s.checkUserAccess(ctx, user, assignmentID)
// 		if err != nil {
// 			return err
// 		}
// 		if !allowed {
// 			return fmt.Errorf("user is not authorized to submit this assignment")
// 		}

// 		// Cek apakah user sudah submit sebelumnya
// 		existing, err := s.submissionRepo.GetByAssignmentUser(ctx, assignmentID, user.UserID)
// 		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 			return err
// 		}
// 		if existing.ID != uuid.Nil {
// 			return fmt.Errorf("submission already exists")
// 		}

// 		submission = models.AssignmentSubmission{
// 			ID:           uuid.New(),
// 			AssignmentID: assignmentID,
// 			UserID:       user.UserID,
// 			Note:         req.Note,
// 			EssayText:    req.EssayText,
// 		}

// 		if err := s.submissionRepo.WithTx(tx).Create(ctx, &submission); err != nil {
// 			return err
// 		}

// 		// Simpan file URLs sebagai SubmissionFile records
// 		var submissionFiles []models.SubmissionFile
// 		for _, url := range fileURLs {
// 			submissionFiles = append(submissionFiles, models.SubmissionFile{
// 				ID:                     uuid.New(),
// 				AssignmentSubmissionID: submission.ID,
// 				FileURL:                url,
// 			})
// 		}

// 		if len(submissionFiles) > 0 {
// 			if err := s.submissionRepo.WithTx(tx).CreateSubmissionFiles(ctx, submissionFiles); err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Ambil submission lengkap dengan files
// 	submission, err = s.submissionRepo.FindByID(ctx, submission.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &submission, nil
// }

// CreateSubmission is for create submission
func (s *SubmissionService) CreateSubmission(
	ctx context.Context,
	user *utils.Claims,
	assignmentID uuid.UUID,
	req *dto.CreateSubmissionRequest,
	fileURLs []string,
) (*models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		// --- 1. Cek apakah user punya akses ke batch ---
		allowed, err := s.checkUserAccess(ctx, user, assignmentID)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("user is not authorized to submit this assignment")
		}

		// --- 2. Cek apakah user sudah pernah submit ---
		existing, err := s.submissionRepo.GetByAssignmentUser(ctx, assignmentID, user.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if existing.ID != uuid.Nil {
			return fmt.Errorf("submission already exists")
		}

		// --- 3. Validasi Meeting Rules ---
		if err := s.validateMeetingRules(ctx, tx, assignmentID, user.UserID); err != nil {
			return err
		}

		// --- 4. Simpan submission utama ---
		submission = models.AssignmentSubmission{
			ID:           uuid.New(),
			AssignmentID: assignmentID,
			UserID:       user.UserID,
			Note:         req.Note,
			EssayText:    req.EssayText,
		}

		if err := s.submissionRepo.WithTx(tx).Create(ctx, &submission); err != nil {
			return err
		}

		// --- 5. Simpan file submissions ---
		var submissionFiles []models.SubmissionFile
		for _, url := range fileURLs {
			submissionFiles = append(submissionFiles, models.SubmissionFile{
				ID:                     uuid.New(),
				AssignmentSubmissionID: submission.ID,
				FileURL:                url,
			})
		}

		if len(submissionFiles) > 0 {
			if err := s.submissionRepo.WithTx(tx).CreateSubmissionFiles(ctx, submissionFiles); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Ambil submission lengkap dengan files
	submission, err = s.submissionRepo.FindByID(ctx, submission.ID)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

func (s *SubmissionService) validateMeetingRules(
	ctx context.Context,
	tx *gorm.DB,
	assignmentID, userID uuid.UUID,
) error {
	assignment, err := s.assignmentRepo.WithTx(tx).FindByID(ctx, assignmentID)
	if err != nil {
		return fmt.Errorf("assignment not found")
	}

	currentMeeting, err := s.meetingRepo.FindByID(ctx, assignment.MeetingID)
	if err != nil {
		return fmt.Errorf("meeting not found")
	}

	if currentMeeting.StartAt.After(currentMeeting.EndAt) {
		return fmt.Errorf("invalid meeting schedule")
	}

	if !currentMeeting.IsOpen {
		return fmt.Errorf("meeting is not open yet")
	}

	// Cek meeting sebelumnya
	prevMeeting, err := s.meetingRepo.WithTx(tx).GetPrevMeeting(ctx, currentMeeting.BatchID, currentMeeting.StartAt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Tidak ada meeting sebelumnya → langsung boleh submit
		return nil
	}
	if err != nil {
		return err
	}

	// --- Validasi kehadiran di meeting sebelumnya ---
	_, err = s.attendanceRepo.GetByMeetingAndUser(ctx, prevMeeting.ID, userID)
	if err != nil {
		// Tidak ada attendance → artinya user tidak hadir, tolak
		return fmt.Errorf("anda belum absen di meeting sebelumnya")
	}
	// if !att.IsPresent {
	// 	return fmt.Errorf("anda tidak hadir di meeting sebelumnya")
	// }

	// --- Validasi assignment di meeting sebelumnya ---
	prevAssignment, err := s.assignmentRepo.GetByMeetingID(ctx, prevMeeting.ID)
	if err == nil {
		// Ada assignment sebelumnya, cek apakah sudah submit
		_, err = s.submissionRepo.FindByAssignmentAndUserID(ctx, prevAssignment.ID, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("anda belum mengumpulkan submission di meeting sebelumnya")
		}
	} // jika assignment tidak ada, skip → boleh submit

	// --- Validasi semua quiz di meeting sebelumnya ---
	prevQuizzes, err := s.quizRepo.WithTx(tx).GetAllByMeetingID(ctx, prevMeeting.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("gagal mengambil quiz meeting sebelumnya")
	}

	for _, q := range prevQuizzes {
		subs, err := s.quizRepo.WithTx(tx).GetQuizSubmissionByQuizAndUser(ctx, q.ID, userID)
		if errors.Is(err, gorm.ErrRecordNotFound) || len(subs) == 0 {
			return fmt.Errorf("anda belum mengerjakan quiz '%s' di meeting sebelumnya", q.Title)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateSubmission for update
func (s *SubmissionService) UpdateSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID, body *dto.UpdateSubmissionRequest) (*models.AssignmentSubmission, error) {
	var updatedSubmission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil submission milik user
		submission, err := s.submissionRepo.WithTx(tx).GetByIDUser(ctx, submissionID, user.UserID)
		if err != nil {
			return err
		}

		// Cek pembayaran
		batchID, err := s.assignmentRepo.GetBatchIDByAssignmentID(ctx, submission.AssignmentID)
		if err != nil {
			return err
		}
		hasPaid, err := s.purchaseService.HasPaid(ctx, user.UserID, batchID)
		if err != nil {
			return err
		}
		if !hasPaid {
			return fmt.Errorf("forbidden: user has not purchased this course")
		}

		// Update data (ignore empty)
		if err := copier.CopyWithOption(&submission, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		if err := s.submissionRepo.WithTx(tx).Update(ctx, submission); err != nil {
			return err
		}

		// Replace files jika dikirim
		if body.SubmissionFiles != nil {
			if err := s.submissionRepo.WithTx(tx).DeleteFilesBySubmissionID(ctx, submission.ID); err != nil {
				return err
			}

			var files []models.SubmissionFile
			for _, f := range *body.SubmissionFiles {
				files = append(files, models.SubmissionFile{
					AssignmentSubmissionID: submission.ID,
					FileURL:                f.FileURL,
				})
			}
			if len(files) > 0 {
				if err := s.submissionRepo.WithTx(tx).CreateFiles(ctx, files); err != nil {
					return err
				}
			}
		}

		// Ambil ulang data untuk response
		fresh, err := s.submissionRepo.WithTx(tx).FindByID(ctx, submission.ID)
		if err != nil {
			return err
		}
		updatedSubmission = utils.Safe(&fresh, models.AssignmentSubmission{})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &updatedSubmission, nil
}

// DeleteSubmission for delete
func (s *SubmissionService) DeleteSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID) error {
	var submission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil submission milik user
		submissionRsp, err := s.submissionRepo.WithTx(tx).GetByIDUser(ctx, submissionID, user.UserID)
		if err != nil {
			return err
		}
		submission = utils.Safe(submissionRsp, models.AssignmentSubmission{})

		// Cek pembayaran
		batchID, err := s.assignmentRepo.GetBatchIDByAssignmentID(ctx, submission.AssignmentID)
		if err != nil {
			return err
		}
		hasPaid, err := s.purchaseService.HasPaid(ctx, user.UserID, batchID)
		if err != nil {
			return err
		}
		if !hasPaid {
			return fmt.Errorf("forbidden: user has not purchased this course")
		}

		// Hapus dari DB
		if err := s.submissionRepo.WithTx(tx).DeleteByID(ctx, submissionID); err != nil {
			return err
		}

		return nil
	})

	// Setelah commit, hapus file di storage
	if len(submission.SubmissionFiles) > 0 {
		for _, f := range submission.SubmissionFiles {
			if err := s.fileService.DeleteFile(f.FileURL); err != nil {
				log.Errorf("Failed to delete file %s: %v", f.FileURL, err)
			}
		}
	}

	return err
}

// GetSubmissionGrade for get submission grade
func (s *SubmissionService) GetSubmissionGrade(ctx context.Context, user *utils.Claims, submissionID uuid.UUID) (*models.AssignmentGrade, error) {
	submission, err := s.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	allowed, err := s.checkUserAccess(ctx, user, submission.AssignmentID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: no access to this assignment")
	}

	if user.Role == string(models.RoleTypeSiswa) && submission.UserID != user.UserID {
		return nil, fmt.Errorf("forbidden: not your submission")
	}

	grade, err := s.submissionRepo.GetGradeBySubmissionID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	return grade, nil
}

// GradeSubmission for post
func (s *SubmissionService) GradeSubmission(ctx context.Context, user *utils.Claims, submissionID uuid.UUID, req *dto.GradeSubmissionRequest) (models.AssignmentGrade, error) {
	submission, err := s.submissionRepo.FindByID(ctx, submissionID)
	if err != nil {
		return models.AssignmentGrade{}, err
	}

	// Pastikan guru punya akses
	allowed, err := s.checkUserAccess(ctx, user, submission.AssignmentID)
	if err != nil {
		return models.AssignmentGrade{}, err
	}
	if !allowed {
		return models.AssignmentGrade{}, fmt.Errorf("forbidden: not teacher of this assignment")
	}

	gradeModel := models.AssignmentGrade{
		AssignmentSubmissionID: submission.ID,
		Grade:                  req.Grade,
		Feedback:               req.Feedback,
		GradedBy:               user.UserID,
	}

	// Upsert nilai
	if _, err := s.submissionRepo.UpsertGrade(ctx, gradeModel); err != nil {
		return models.AssignmentGrade{}, err
	}

	// Ambil lagi grade yang sudah tersimpan
	grade, err := s.submissionRepo.GetGradeBySubmissionID(ctx, submissionID)
	if err != nil {
		return models.AssignmentGrade{}, err
	}

	return *grade, nil
}

// GenerateGradesExcel services
func (s *SubmissionService) GenerateGradesExcel(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (*excelize.File, string, error) {
	// Cek akses guru
	allowed, err := s.checkUserAccess(ctx, user, assignmentID)
	if err != nil {
		return nil, "", err
	}
	if !allowed {
		return nil, "", fmt.Errorf("forbidden: not teacher of this assignment")
	}

	// Ambil data submission + grade
	submissions, err := s.submissionRepo.GetGradesByAssignmentID(ctx, assignmentID)
	if err != nil {
		return nil, "", err
	}

	// Buat file Excel
	f := excelize.NewFile()
	sheet := "Penilaian"
	f.SetSheetName("Sheet1", sheet)

	// Header
	headers := []string{"UserID", "No", "Nama Siswa", "Nilai", "Feedback"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Isi data
	for i, sub := range submissions {
		no := i + 1
		userID := sub.UserID.String()
		name := sub.User.Name
		grade := ""
		feedback := ""
		if sub.AssignmentGrade != nil {
			grade = fmt.Sprintf("%d", sub.AssignmentGrade.Grade)
			feedback = sub.AssignmentGrade.Feedback
		}
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), userID) // hidden column
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), no)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), grade)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), feedback)
	}

	// Sembunyikan kolom UserID
	f.SetColVisible(sheet, "A", false)

	// Nama file
	filename := fmt.Sprintf("penilaian_%s.xlsx", assignmentID.String())

	return f, filename, nil
}

// ImportGradesFromExcel excel
func (s *SubmissionService) ImportGradesFromExcel(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, fileHeader *multipart.FileHeader) error {
	// Cek akses guru
	allowed, err := s.checkUserAccess(ctx, user, assignmentID)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("forbidden: not teacher of this assignment")
	}

	// Buka file Excel
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

	// Loop data mulai dari row 2 (skip header)
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) < 5 {
			continue // skip kalau kolom tidak lengkap
		}

		userIDStr := rows[i][0] // hidden column
		gradeStr := rows[i][3]
		feedback := rows[i][4]

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fmt.Errorf("row %d: invalid UserID", i+1)
		}

		grade, _ := strconv.Atoi(gradeStr)

		// Cari submission ID berdasarkan assignmentID + userID
		submission, err := s.submissionRepo.FindByAssignmentAndUserID(ctx, assignmentID, userID)
		if err != nil {
			return fmt.Errorf("row %d: %v", i+1, err)
		}

		// Upsert nilai
		gradeModel := models.AssignmentGrade{
			AssignmentSubmissionID: submission.ID,
			Grade:                  grade,
			Feedback:               feedback,
			GradedBy:               user.UserID,
		}
		if _, err := s.submissionRepo.UpsertGrade(ctx, gradeModel); err != nil {
			return fmt.Errorf("row %d: %v", i+1, err)
		}
	}

	return nil
}
