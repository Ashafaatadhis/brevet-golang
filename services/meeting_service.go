package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// MeetingService provides methods for managing meetings
type MeetingService struct {
	meetingRepo  *repository.MeetingRepository
	batchRepo    *repository.BatchRepository
	purchaseRepo *repository.PurchaseRepository
	userRepo     repository.IUserTXRepository
	db           *gorm.DB
}

// NewMeetingService creates a new instance of MeetingService
func NewMeetingService(meetingRepo *repository.MeetingRepository, batchRepo *repository.BatchRepository, purchaseRepo *repository.PurchaseRepository, userRepo repository.IUserTXRepository, db *gorm.DB) *MeetingService {
	return &MeetingService{meetingRepo: meetingRepo, batchRepo: batchRepo, purchaseRepo: purchaseRepo, userRepo: userRepo, db: db}
}

// GetAllFilteredMeetings retrieves all meetings with pagination and filtering options
func (s *MeetingService) GetAllFilteredMeetings(opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	meetings, total, err := s.meetingRepo.GetAllFilteredMeetings(opts)
	if err != nil {
		return nil, 0, err
	}
	return meetings, total, nil
}

// GetMeetingsByBatchSlug retrieves all meetings with pagination and filtering options
func (s *MeetingService) GetMeetingsByBatchSlug(batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	meetings, total, err := s.meetingRepo.GetMeetingsByBatchSlugFiltered(batchSlug, opts)
	if err != nil {
		return nil, 0, err
	}
	return meetings, total, nil
}

// GetMeetingByID retrieves a meeting by its id
func (s *MeetingService) GetMeetingByID(id uuid.UUID) (*models.Meeting, error) {
	meeting, err := s.meetingRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return meeting, nil
}

// CreateMeeting creates a new meeting with the provided details
func (s *MeetingService) CreateMeeting(batchID uuid.UUID, body *dto.CreateMeetingRequest) (*models.Meeting, error) {
	var meeting *models.Meeting

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		_, err := s.batchRepo.WithTx(tx).FindByID(batchID)
		if err != nil {
			return err
		}

		meeting = &models.Meeting{
			ID:          uuid.New(),
			BatchID:     batchID,
			Title:       body.Title,
			Description: body.Description,
			Type:        models.MeetingType(body.Type),
		}

		if err := s.meetingRepo.WithTx(tx).Create(meeting); err != nil {
			return err
		}

		// ✅ Ambil ulang dari DB untuk dapet semua kolom yang terisi otomatis (CreatedAt, dll)
		updated, err := s.meetingRepo.WithTx(tx).FindByID(meeting.ID)
		if err != nil {
			return err
		}
		meeting = updated

		return nil
	})

	if err != nil {
		return nil, err
	}
	return meeting, nil
}

// UpdateMeeting updates an existing meeting with the provided details
func (s *MeetingService) UpdateMeeting(batchID uuid.UUID, body *dto.UpdateMeetingRequest) (*models.Meeting, error) {
	var meeting models.Meeting
	fmt.Println(&body, " TTT")

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		meetingPtr, err := s.meetingRepo.WithTx(tx).FindByID(batchID)
		if err != nil {
			return fmt.Errorf("meeting tidak ditemukan: %w", err)
		}

		meeting = utils.Safe(meetingPtr, models.Meeting{})

		if err := copier.Copy(&meeting, &body); err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		if err := s.meetingRepo.WithTx(tx).Update(&meeting); err != nil {
			return err
		}

		// ✅ Ambil ulang dari DB untuk dapet semua kolom yang terisi otomatis (CreatedAt, dll)
		updated, err := s.meetingRepo.WithTx(tx).FindByID(meeting.ID)
		if err != nil {
			return err
		}
		meeting = utils.Safe(updated, models.Meeting{})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

// DeleteMeeting deletes a meeting by its ID
func (s *MeetingService) DeleteMeeting(id uuid.UUID) error {

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		var err error
		_, err = s.meetingRepo.WithTx(tx).FindByID(id)
		if err != nil {
			return err
		}

		// Hapus meeting
		if err := s.meetingRepo.WithTx(tx).DeleteByID(id); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}
	return nil

}

// AddTeachersToMeeting is service for add teacher to meeting
func (s *MeetingService) AddTeachersToMeeting(meetingID uuid.UUID, req *dto.AssignTeachersRequest) (*models.Meeting, error) {
	var meeting models.Meeting
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		// Ambil semua guru berdasarkan ID yang diminta
		teachers, err := s.userRepo.WithTx(tx).FindByIDs(req.TeacherIDs)
		if err != nil {
			return err
		}

		if len(teachers) != len(req.TeacherIDs) {
			return fmt.Errorf("beberapa teacher_id tidak ditemukan")
		}

		// Validasi role guru
		for _, t := range teachers {
			if t.RoleType != models.RoleTypeGuru {
				return fmt.Errorf("user dengan ID %s bukan guru", t.ID.String())
			}
		}

		// Ambil semua user_id guru yang sudah tergabung di meeting ini
		existingIDs, err := s.meetingRepo.WithTx(tx).GetTeacherIDsByMeetingID(meetingID)
		if err != nil {
			return err
		}

		// Buat map untuk cek cepat
		existingSet := make(map[uuid.UUID]bool)
		for _, id := range existingIDs {
			existingSet[id] = true
		}

		// Filter guru baru yang belum tergabung
		var newTeacherIDs []uuid.UUID
		for _, id := range req.TeacherIDs {
			if !existingSet[id] {
				newTeacherIDs = append(newTeacherIDs, id)
			}
		}

		if len(newTeacherIDs) == 0 {
			return fmt.Errorf("semua guru yang dikirim sudah tergabung dalam meeting ini")
		}

		if len(newTeacherIDs) < len(req.TeacherIDs) {
			var skipped []string
			for _, id := range req.TeacherIDs {
				if existingSet[id] {
					skipped = append(skipped, id.String())
				}
			}
			return fmt.Errorf("beberapa guru sudah tergabung: %v", skipped)
		}

		// Tambahkan guru baru
		meetingResp, err := s.meetingRepo.WithTx(tx).AddTeachers(meetingID, newTeacherIDs)
		if err != nil {
			return err
		}

		meeting = utils.Safe(meetingResp, models.Meeting{})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &meeting, nil

}

// UpdateTeachersInMeeting this function service for update teacher in meeting
func (s *MeetingService) UpdateTeachersInMeeting(meetingID uuid.UUID, req *dto.AssignTeachersRequest) (*models.Meeting, error) {
	var meeting models.Meeting
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil user dari DB
		teachers, err := s.userRepo.WithTx(tx).FindByIDs(req.TeacherIDs)
		if err != nil {
			return err
		}

		// Cek jumlah cocok
		if len(teachers) != len(req.TeacherIDs) {
			return fmt.Errorf("beberapa teacher_id tidak ditemukan")
		}

		// Validasi role guru
		for _, t := range teachers {
			if t.RoleType != models.RoleTypeGuru {
				return fmt.Errorf("user dengan ID %s bukan guru", t.ID.String())
			}
		}

		// Update guru di meeting
		meetingPtr, err := s.meetingRepo.WithTx(tx).UpdateTeachers(meetingID, req.TeacherIDs)
		if err != nil {
			return err
		}

		meeting = utils.Safe(meetingPtr, models.Meeting{})

		return nil

	})

	if err != nil {
		return nil, err
	}

	return &meeting, nil
}

// RemoveTeachersFromMeeting this function repo for remove teacher
func (s *MeetingService) RemoveTeachersFromMeeting(meetingID uuid.UUID, teacherID uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil user dari DB
		teachers, err := s.userRepo.WithTx(tx).FindByID(teacherID)
		if err != nil {
			return err
		}

		// Validasi role guru
		if teachers.RoleType != models.RoleTypeGuru {
			return fmt.Errorf("user dengan ID %s bukan guru", teachers.ID.String())
		}

		// Hapus guru dari meeting
		meetingPtr, err := s.meetingRepo.WithTx(tx).RemoveTeacher(meetingID, teacherID)
		if err != nil {
			return err
		}

		meeting = utils.Safe(meetingPtr, models.Meeting{})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &meeting, nil

}

// GetTeachersByMeetingIDFiltered is function get all teachers by meeting id
func (s *MeetingService) GetTeachersByMeetingIDFiltered(meetingID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	meetings, total, err := s.meetingRepo.GetTeachersByMeetingIDFiltered(meetingID, opts)
	if err != nil {
		return nil, 0, err
	}
	return meetings, total, nil
}

// GetStudentsByBatchSlugFiltered for get all students by batch
func (s *MeetingService) GetStudentsByBatchSlugFiltered(user *utils.Claims, batchSlug string, opts utils.QueryOptions) ([]models.User, int64, error) {
	if models.RoleType(user.Role) == models.RoleTypeAdmin {
		return s.meetingRepo.GetStudentsByBatchSlugFiltered(batchSlug, opts)
	}

	owned, err := s.meetingRepo.IsBatchOwnedByUser(user.UserID, batchSlug)
	if err != nil {
		return nil, 0, err
	}
	if !owned {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, "Anda tidak punya akses ke batch ini")
	}

	return s.meetingRepo.GetStudentsByBatchSlugFiltered(batchSlug, opts)
}

// GetMeetingsPurchasedByUser is service for get meetings where the user has purchased
func (s *MeetingService) GetMeetingsPurchasedByUser(userID uuid.UUID, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	batch, err := s.batchRepo.GetBatchBySlug(batchSlug)
	if err != nil {
		return nil, 0, err
	}

	hasPaid, err := s.purchaseRepo.HasPaid(userID, batch.ID)
	if err != nil {
		return nil, 0, err
	}
	if !hasPaid {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, "Anda belum membeli batch ini")
	}

	return s.meetingRepo.GetMeetingsByBatchSlugFiltered(batchSlug, opts)
}

// GetMeetingsTaughtByTeacher is service for get meetings where the teacher has taught
func (s *MeetingService) GetMeetingsTaughtByTeacher(userID uuid.UUID, batchSlug string, opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	// 🔒 Validasi kepemilikan batch
	owned, err := s.meetingRepo.IsBatchOwnedByUser(userID, batchSlug)
	if err != nil {
		return nil, 0, err
	}
	if !owned {
		return nil, 0, fiber.NewError(fiber.StatusForbidden, "Anda tidak mengajar di batch ini")
	}

	return s.meetingRepo.GetMeetingsByBatchSlugFiltered(batchSlug, opts)
}
