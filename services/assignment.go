package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignmentService provides methods for managing assignments
type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	meetingRepo    *repository.MeetingRepository
	db             *gorm.DB
}

// NewAssignmentService creates a new instance of AssignmentService
func NewAssignmentService(assignmentRepository *repository.AssignmentRepository, meetingRepository *repository.MeetingRepository, db *gorm.DB) *AssignmentService {
	return &AssignmentService{assignmentRepo: assignmentRepository, meetingRepo: meetingRepository, db: db}
}

// GetAllFilteredAssignments retrieves all assignments with pagination and filtering options
func (s *AssignmentService) GetAllFilteredAssignments(opts utils.QueryOptions) ([]models.Assignment, int64, error) {
	assignments, total, err := s.assignmentRepo.GetAllFilteredAssignments(opts)
	if err != nil {
		return nil, 0, err
	}
	return assignments, total, nil
}

// CreateAssignment creates a new assignment with the provided details
func (s *AssignmentService) CreateAssignment(user *utils.Claims, meetingID uuid.UUID, body *dto.CreateAssignmentRequest) (*models.Assignment, error) {
	var assignment models.Assignment

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		meeting, err := s.meetingRepo.WithTx(tx).FindByID(meetingID)
		if err != nil {
			return err
		}

		// 🛡️ Access Control: hanya guru yang bersangkutan atau admin yang boleh create
		if user.Role == string(models.RoleTypeGuru) && meeting.TeacherID != user.UserID {
			return fmt.Errorf("forbidden: user %s not authorized to create assignment for meeting %s", user.UserID, meeting.ID)
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

		if err := s.assignmentRepo.WithTx(tx).Create(assignmentPtr); err != nil {
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
			if err := s.assignmentRepo.WithTx(tx).CreateFiles(files); err != nil {
				return err
			}
		}

		// ✅ Ambil ulang dari DB untuk dapet semua kolom yang terisi otomatis (CreatedAt, dll)
		updated, err := s.assignmentRepo.WithTx(tx).FindByID(assignmentPtr.ID)
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
