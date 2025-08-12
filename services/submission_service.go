package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// SubmissionService provides methods for managing submissions
type SubmissionService struct {
	submissionRepo  *repository.SubmissionRepository
	assignmentRepo  *repository.AssignmentRepository
	meetingRepo     *repository.MeetingRepository
	purchaseService *PurchaseService
	fileService     *FileService
	db              *gorm.DB
}

// NewSubmissionService creates a new instance of SubmissionService
func NewSubmissionService(submissionRepo *repository.SubmissionRepository, assignmentRepo *repository.AssignmentRepository, meetingRepo *repository.MeetingRepository, purchaseService *PurchaseService, fileService *FileService, db *gorm.DB) *SubmissionService {
	return &SubmissionService{submissionRepo: submissionRepo, assignmentRepo: assignmentRepo, meetingRepo: meetingRepo, purchaseService: purchaseService, fileService: fileService, db: db}
}

func (s *SubmissionService) checkUserAccess(user *utils.Claims, assignmentID uuid.UUID) (bool, error) {
	// Cari batch info dari assignmentID
	batch, err := s.assignmentRepo.GetBatchByAssignmentID(assignmentID) // balikin batchSlug & batchID
	if err != nil {
		return false, err
	}

	// Kalau role teacher, cek apakah dia mengajar batch ini
	if user.Role == string(models.RoleTypeGuru) {
		return s.meetingRepo.IsBatchOwnedByUser(user.UserID, batch.Slug)
	}

	// Kalau student, cek pembayaran
	if user.Role == string(models.RoleTypeSiswa) {
		return s.purchaseService.HasPaid(user.UserID, batch.ID)
	}

	// Role lain tidak diizinkan
	return false, nil
}

// GetAllSubmissionsByAssignmentUser for get all
func (s *SubmissionService) GetAllSubmissionsByAssignmentUser(assignmentID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.AssignmentSubmission, int64, error) {
	allowed, err := s.checkUserAccess(user, assignmentID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, errors.New("user not authorized to access this assignment")
	}
	if user.Role == string(models.RoleTypeGuru) {
		return s.submissionRepo.GetAllByAssignment(assignmentID, nil, opts)
	}
	return s.submissionRepo.GetAllByAssignment(assignmentID, &user.UserID, opts)

}

// GetSubmissionDetail fot get detail
func (s *SubmissionService) GetSubmissionDetail(submissionID uuid.UUID, user *utils.Claims) (models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission
	var err error

	if user.Role == string(models.RoleTypeGuru) {
		submission, err = s.submissionRepo.FindByID(submissionID)
	} else {
		subPtr, err2 := s.submissionRepo.GetByIDUser(submissionID, user.UserID)
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
	allowed, err := s.checkUserAccess(user, submission.AssignmentID)
	if err != nil {
		return models.AssignmentSubmission{}, err
	}
	if !allowed {
		return models.AssignmentSubmission{}, errors.New("user not authorized to access this assignment")
	}

	return submission, nil
}

// CreateSubmission is for create submission
func (s *SubmissionService) CreateSubmission(user *utils.Claims, assignmentID uuid.UUID, note string, fileURLs []string) (*models.AssignmentSubmission, error) {
	var submission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Cek apakah user sudah bayar batch terkait assignment
		allowed, err := s.checkUserAccess(user, assignmentID)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("user is not authorized to submit this assignment")
		}

		// Cek apakah user sudah submit sebelumnya
		existing, err := s.submissionRepo.GetByAssignmentUser(assignmentID, user.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if existing.ID != uuid.Nil {
			return fmt.Errorf("submission already exists")
		}

		submission = models.AssignmentSubmission{
			ID:           uuid.New(),
			AssignmentID: assignmentID,
			UserID:       user.UserID,
			Note:         note,
			SubmittedAt:  time.Now(),
			IsLate:       false, // opsional: bisa cek assignment.EndAt dibanding now
		}

		if err := s.submissionRepo.WithTx(tx).Create(&submission); err != nil {
			return err
		}

		// Simpan file URLs sebagai SubmissionFile records
		var submissionFiles []models.SubmissionFile
		for _, url := range fileURLs {
			submissionFiles = append(submissionFiles, models.SubmissionFile{
				ID:                     uuid.New(),
				AssignmentSubmissionID: submission.ID,
				FileURL:                url,
			})
		}

		if len(submissionFiles) > 0 {
			if err := s.submissionRepo.WithTx(tx).CreateSubmissionFiles(submissionFiles); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Ambil submission lengkap dengan files
	submission, err = s.submissionRepo.FindByID(submission.ID)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// UpdateSubmission for update
func (s *SubmissionService) UpdateSubmission(user *utils.Claims, submissionID uuid.UUID, body *dto.UpdateSubmissionRequest) (*models.AssignmentSubmission, error) {
	var updatedSubmission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil submission milik user
		submission, err := s.submissionRepo.WithTx(tx).GetByIDUser(submissionID, user.UserID)
		if err != nil {
			return err
		}

		// Cek pembayaran
		batchID, err := s.assignmentRepo.GetBatchIDByAssignmentID(submission.AssignmentID)
		if err != nil {
			return err
		}
		hasPaid, err := s.purchaseService.HasPaid(user.UserID, batchID)
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

		if err := s.submissionRepo.WithTx(tx).Update(submission); err != nil {
			return err
		}

		// Replace files jika dikirim
		if body.SubmissionFiles != nil {
			if err := s.submissionRepo.WithTx(tx).DeleteFilesBySubmissionID(submission.ID); err != nil {
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
				if err := s.submissionRepo.WithTx(tx).CreateFiles(files); err != nil {
					return err
				}
			}
		}

		// Ambil ulang data untuk response
		fresh, err := s.submissionRepo.WithTx(tx).FindByID(submission.ID)
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
func (s *SubmissionService) DeleteSubmission(user *utils.Claims, submissionID uuid.UUID) error {
	var submission models.AssignmentSubmission

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil submission milik user
		submissionRsp, err := s.submissionRepo.WithTx(tx).GetByIDUser(submissionID, user.UserID)
		if err != nil {
			return err
		}
		submission = utils.Safe(submissionRsp, models.AssignmentSubmission{})

		// Cek pembayaran
		batchID, err := s.assignmentRepo.GetBatchIDByAssignmentID(submission.AssignmentID)
		if err != nil {
			return err
		}
		hasPaid, err := s.purchaseService.HasPaid(user.UserID, batchID)
		if err != nil {
			return err
		}
		if !hasPaid {
			return fmt.Errorf("forbidden: user has not purchased this course")
		}

		// Hapus dari DB
		if err := s.submissionRepo.WithTx(tx).DeleteByID(submissionID); err != nil {
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
func (s *SubmissionService) GetSubmissionGrade(user *utils.Claims, submissionID uuid.UUID) (*models.AssignmentGrade, error) {
	submission, err := s.submissionRepo.FindByID(submissionID)
	if err != nil {
		return nil, err
	}

	allowed, err := s.checkUserAccess(user, submission.AssignmentID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden: no access to this assignment")
	}

	if user.Role == string(models.RoleTypeSiswa) && submission.UserID != user.UserID {
		return nil, fmt.Errorf("forbidden: not your submission")
	}

	grade, err := s.submissionRepo.GetGradeBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	return grade, nil
}

// GradeSubmission for post
func (s *SubmissionService) GradeSubmission(user *utils.Claims, submissionID uuid.UUID, req *dto.GradeSubmissionRequest) (models.AssignmentGrade, error) {
	submission, err := s.submissionRepo.FindByID(submissionID)
	if err != nil {
		return models.AssignmentGrade{}, err
	}

	// Pastikan guru punya akses
	allowed, err := s.checkUserAccess(user, submission.AssignmentID)
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
	if _, err := s.submissionRepo.UpsertGrade(gradeModel); err != nil {
		return models.AssignmentGrade{}, err
	}

	// Ambil lagi grade yang sudah tersimpan
	grade, err := s.submissionRepo.GetGradeBySubmissionID(submissionID)
	if err != nil {
		return models.AssignmentGrade{}, err
	}

	return *grade, nil
}
