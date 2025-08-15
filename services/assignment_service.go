package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"context"

	"brevet-api/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IAssignmentService interface
type IAssignmentService interface {
	GetAllFilteredAssignments(ctx context.Context, opts utils.QueryOptions) ([]models.Assignment, int64, error)
	GetAllFilteredAssignmentsByMeetingID(ctx context.Context, meetingID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.Assignment, int64, error)
	GetAssignmentByID(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (*models.Assignment, error)
	CreateAssignment(ctx context.Context, user *utils.Claims, meetingID uuid.UUID, body *dto.CreateAssignmentRequest) (*models.Assignment, error)
	UpdateAssignment(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, body *dto.UpdateAssignmentRequest) (*models.Assignment, error)
	DeleteAssignment(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) error
}

// AssignmentService provides methods for managing assignments
type AssignmentService struct {
	assignmentRepo repository.IAssignmentRepository
	meetingRepo    repository.IMeetingRepository
	purchaseRepo   repository.IPurchaseRepository
	fileService    IFileService
	db             *gorm.DB
}

// NewAssignmentService creates a new instance of AssignmentService
func NewAssignmentService(assignmentRepository repository.IAssignmentRepository, meetingRepository repository.IMeetingRepository,
	purchaseRepo repository.IPurchaseRepository, fileService IFileService, db *gorm.DB) IAssignmentService {
	return &AssignmentService{assignmentRepo: assignmentRepository, meetingRepo: meetingRepository, purchaseRepo: purchaseRepo, fileService: fileService, db: db}
}

// GetAllFilteredAssignments retrieves all assignments with pagination and filtering options
func (s *AssignmentService) GetAllFilteredAssignments(ctx context.Context, opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	assignments, total, err := s.assignmentRepo.GetAllFilteredAssignments(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	return assignments, total, nil
}

// GetAllFilteredAssignmentsByMeetingID retrieves all assignments with pagination and filtering options
func (s *AssignmentService) GetAllFilteredAssignmentsByMeetingID(ctx context.Context, meetingID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	if user.Role == string(models.RoleTypeGuru) {
		ok, err := s.meetingRepo.IsMeetingTaughtByUser(ctx, meetingID, user.UserID)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, err
		}
	}

	assignments, total, err := s.assignmentRepo.GetAllFilteredAssignmentsByMeetingID(ctx, meetingID, opts)
	if err != nil {
		return nil, 0, err
	}
	return assignments, total, nil
}

// GetAssignmentByID retrieves a single assignment by its ID
func (s *AssignmentService) GetAssignmentByID(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) (*models.Assignment, error) {
	assignment, err := s.assignmentRepo.FindByIDWithUserData(ctx, assignmentID, models.RoleType(user.Role), user.UserID)
	if err != nil {
		return nil, err
	}

	switch user.Role {
	case string(models.RoleTypeAdmin):
		// ✅ Admin bebas ambil
		return assignment, nil

	case string(models.RoleTypeGuru):
		// 🔒 Guru hanya boleh jika ngajar di meeting terkait
		isGuru, err := s.meetingRepo.IsUserTeachingInMeeting(ctx, user.UserID, assignment.MeetingID)
		if err != nil {
			return nil, err
		}
		if !isGuru {
			return nil, fmt.Errorf("Anda bukan pengajar di meeting ini")
		}
		return assignment, nil

	case string(models.RoleTypeSiswa):
		meeting, err := s.meetingRepo.FindByID(ctx, assignment.MeetingID)
		if err != nil {
			return nil, err
		}
		// 🔒 Siswa hanya bisa jika sudah beli batch meeting tersebut
		isPurchased, err := s.purchaseRepo.HasPaid(ctx, user.UserID, meeting.BatchID)
		if err != nil {
			return nil, err
		}
		if !isPurchased {
			return nil, fiber.NewError(fiber.StatusForbidden, "Anda belum terdaftar di batch meeting ini")
		}
		return assignment, nil

	default:
		return nil, fiber.NewError(fiber.StatusForbidden, "Role tidak dikenali")
	}
}

// CreateAssignment creates a new assignment with the provided details
func (s *AssignmentService) CreateAssignment(ctx context.Context, user *utils.Claims, meetingID uuid.UUID, body *dto.CreateAssignmentRequest) (*models.Assignment, error) {
	var assignment models.Assignment

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		meeting, err := s.meetingRepo.WithTx(tx).FindByID(ctx, meetingID)
		if err != nil {
			return err
		}

		// 🛡️ Access Control: hanya guru yang bersangkutan atau admin yang boleh create
		// if user.Role == string(models.RoleTypeGuru) && meeting.TeacherID != user.UserID {
		// 	return fmt.Errorf("forbidden: user %s not authorized to create assignment for meeting %s", user.UserID, meeting.ID)
		// }

		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(ctx, meeting.ID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, meeting.ID)
			}
		}

		assignmentPtr := &models.Assignment{
			ID:          uuid.New(),
			TeacherID:   user.UserID,
			MeetingID:   meetingID,
			StartAt:     body.StartAt,
			EndAt:       body.EndAt,
			Title:       body.Title,
			Description: utils.SafeNil(body.Description),
			Type:        models.AssignmentType(body.Type),
		}

		if err := s.assignmentRepo.WithTx(tx).Create(ctx, assignmentPtr); err != nil {
			return err
		}

		var files []models.AssignmentFiles
		for _, f := range body.AssignmentFiles {
			files = append(files, models.AssignmentFiles{
				AssignmentID: assignmentPtr.ID,
				FileURL:      f,
			})
		}

		if len(files) > 0 {
			if err := s.assignmentRepo.WithTx(tx).CreateFiles(ctx, files); err != nil {
				return err
			}
		}

		// ✅ Ambil ulang dari DB untuk dapet semua kolom yang terisi otomatis (CreatedAt, dll)
		updated, err := s.assignmentRepo.WithTx(tx).FindByID(ctx, assignmentPtr.ID)
		if err != nil {
			return err
		}
		assignment = utils.Safe(updated, models.Assignment{})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// UpdateAssignment updates an existing assignment and its files
func (s *AssignmentService) UpdateAssignment(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID, body *dto.UpdateAssignmentRequest) (*models.Assignment, error) {
	var updatedAssignment models.Assignment

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		assignment, err := s.assignmentRepo.WithTx(tx).FindByID(ctx, assignmentID)
		if err != nil {
			return err
		}

		// 🛡️ Access Control: hanya guru pemilik atau admin yang bisa update
		// if user.Role == string(models.RoleTypeGuru) && assignment.TeacherID != user.UserID {
		// 	return fmt.Errorf("forbidden: user %s not authorized to update assignment %s", user.UserID, assignment.ID)
		// }
		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(ctx, assignment.MeetingID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, assignment.MeetingID)
			}
		}

		// Copy field yang tidak nil saja
		if err := copier.CopyWithOption(&assignment, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		if err := s.assignmentRepo.WithTx(tx).Update(ctx, assignment); err != nil {
			return err
		}

		// Optional: replace files (delete old, insert new)
		if body.AssignmentFiles != nil {
			// Hapus semua file lama
			if err := s.assignmentRepo.WithTx(tx).DeleteFilesByAssignmentID(ctx, assignment.ID); err != nil {
				return err
			}

			// Masukkan file baru
			var files []models.AssignmentFiles
			for _, f := range body.AssignmentFiles {
				files = append(files, models.AssignmentFiles{
					AssignmentID: assignment.ID,
					FileURL:      f,
				})
			}

			if len(files) > 0 {
				if err := s.assignmentRepo.WithTx(tx).CreateFiles(ctx, files); err != nil {
					return err
				}
			}
		}

		// Ambil ulang assignment lengkap
		fresh, err := s.assignmentRepo.WithTx(tx).FindByID(ctx, assignment.ID)
		if err != nil {
			return err
		}
		updatedAssignment = utils.Safe(fresh, models.Assignment{})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &updatedAssignment, nil
}

// DeleteAssignment deletes an assignment and its related files
func (s *AssignmentService) DeleteAssignment(ctx context.Context, user *utils.Claims, assignmentID uuid.UUID) error {
	var assignment models.Assignment

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		assignmentRsp, err := s.assignmentRepo.WithTx(tx).FindByID(ctx, assignmentID)
		if err != nil {
			return err
		}

		// 🛡️ Access Control
		// if user.Role == string(models.RoleTypeGuru) && assignmentRsp.TeacherID != user.UserID {
		// 	return fmt.Errorf("forbidden: user %s not authorized to delete assignment %s", user.UserID, assignmentRsp.ID)
		// }
		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(ctx, assignmentRsp.MeetingID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, assignmentRsp.MeetingID)
			}
		}

		assignment = utils.Safe(assignmentRsp, models.Assignment{})

		// Hapus dari DB (files ikut kehapus karena CASCADE)
		if err := s.assignmentRepo.WithTx(tx).DeleteByID(ctx, assignmentID); err != nil {
			return err
		}

		return nil
	})

	// Setelah commit, hapus file dari cloud atau disk
	if len(assignment.AssignmentFiles) > 0 {
		for _, f := range assignment.AssignmentFiles {
			if err := s.fileService.DeleteFile(f.FileURL); err != nil {
				log.Errorf("Gagal hapus file %s: %v", f.FileURL, err)
			}
		}
	}

	if err != nil {
		return err
	}
	return nil
}
