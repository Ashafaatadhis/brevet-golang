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

// MeetingService provides methods for managing meetings
type MeetingService struct {
	meetingRepo *repository.MeetingRepository
	userRepo    *repository.UserRepository
	db          *gorm.DB
}

// NewMeetingService creates a new instance of MeetingService
func NewMeetingService(meetingRepo *repository.MeetingRepository, userRepo *repository.UserRepository, db *gorm.DB) *MeetingService {
	return &MeetingService{meetingRepo: meetingRepo, userRepo: userRepo, db: db}
}

// GetAllFilteredMeetings retrieves all meetings with pagination and filtering options
func (s *MeetingService) GetAllFilteredMeetings(opts utils.QueryOptions) ([]models.Meeting, int64, error) {
	meetings, total, err := s.meetingRepo.GetAllFilteredMeetings(opts)
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
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	meeting := &models.Meeting{
		ID:          uuid.New(),
		BatchID:     batchID,
		Title:       body.Title,
		Description: body.Description,
		Type:        models.MeetingType(body.Type),
	}

	if err := s.meetingRepo.WithTx(tx).Create(meeting); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return meeting, nil
}

// AddTeachersToMeeting is service for add teacher to meeting
func (s *MeetingService) AddTeachersToMeeting(meetingID uuid.UUID, req *dto.AssignTeachersRequest) (*models.Meeting, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Ambil semua guru berdasarkan ID yang diminta
	teachers, err := s.userRepo.WithTx(tx).FindByIDs(req.TeacherIDs)
	if err != nil {
		return nil, err
	}

	if len(teachers) != len(req.TeacherIDs) {
		return nil, fmt.Errorf("beberapa teacher_id tidak ditemukan")
	}

	// Validasi role guru
	for _, t := range teachers {
		if t.RoleType != models.RoleTypeGuru {
			return nil, fmt.Errorf("user dengan ID %s bukan guru", t.ID.String())
		}
	}

	// Ambil semua user_id guru yang sudah tergabung di meeting ini
	existingIDs, err := s.meetingRepo.WithTx(tx).GetTeacherIDsByMeetingID(meetingID)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("semua guru yang dikirim sudah tergabung dalam meeting ini")
	}

	if len(newTeacherIDs) < len(req.TeacherIDs) {
		var skipped []string
		for _, id := range req.TeacherIDs {
			if existingSet[id] {
				skipped = append(skipped, id.String())
			}
		}
		return nil, fmt.Errorf("beberapa guru sudah tergabung: %v", skipped)
	}

	// Tambahkan guru baru
	meeting, err := s.meetingRepo.WithTx(tx).AddTeachers(meetingID, newTeacherIDs)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return meeting, nil
}

// UpdateTeachersInMeeting this function service for update teacher in meeting
func (s *MeetingService) UpdateTeachersInMeeting(meetingID uuid.UUID, req *dto.AssignTeachersRequest) (*models.Meeting, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Ambil user dari DB
	teachers, err := s.userRepo.WithTx(tx).FindByIDs(req.TeacherIDs)
	if err != nil {
		return nil, err
	}

	// Cek jumlah cocok
	if len(teachers) != len(req.TeacherIDs) {
		return nil, fmt.Errorf("beberapa teacher_id tidak ditemukan")
	}

	// Validasi role guru
	for _, t := range teachers {
		if t.RoleType != models.RoleTypeGuru {
			return nil, fmt.Errorf("user dengan ID %s bukan guru", t.ID.String())
		}
	}

	// Update guru di meeting
	meeting, err := s.meetingRepo.WithTx(tx).UpdateTeachers(meetingID, req.TeacherIDs)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return meeting, nil
}

// RemoveTeachersFromMeeting this function repo for remove teacher
func (s *MeetingService) RemoveTeachersFromMeeting(meetingID uuid.UUID, teacherID uuid.UUID) (*models.Meeting, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Ambil user dari DB
	teachers, err := s.userRepo.WithTx(tx).FindByID(teacherID)
	if err != nil {
		return nil, err
	}

	// Validasi role guru
	if teachers.RoleType != models.RoleTypeGuru {
		return nil, fmt.Errorf("user dengan ID %s bukan guru", teachers.ID.String())
	}

	// Hapus guru dari meeting
	meeting, err := s.meetingRepo.WithTx(tx).RemoveTeacher(meetingID, teacherID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return meeting, nil
}

// GetTeachersByMeetingIDFiltered is function get all teachers by meeting id
func (s *MeetingService) GetTeachersByMeetingIDFiltered(meetingID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	meetings, total, err := s.meetingRepo.GetTeachersByMeetingIDFiltered(meetingID, opts)
	if err != nil {
		return nil, 0, err
	}
	return meetings, total, nil
}
