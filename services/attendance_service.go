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

// AttendanceServices provides methods for managing assignments
type AttendanceServices struct {
	attendanceRepo *repository.AttendanceRepository
	meetingRepo    *repository.MeetingRepository
	purchaseRepo   *repository.PurchaseRepository
	db             *gorm.DB
}

// NewAttendanceService creates a new instance of AssignmentService
func NewAttendanceService(attendanceRepo *repository.AttendanceRepository, meetingRepo *repository.MeetingRepository,
	purchaseRepo *repository.PurchaseRepository, db *gorm.DB) *AttendanceServices {
	return &AttendanceServices{attendanceRepo: attendanceRepo, meetingRepo: meetingRepo, purchaseRepo: purchaseRepo, db: db}
}

// GetAllFilteredAttendances retrieves all attendances with pagination and filtering options
func (s *AttendanceServices) GetAllFilteredAttendances(opts utils.QueryOptions) ([]models.Attendance, int64, error) {
	attendances, total, err := s.attendanceRepo.GetAllFilteredAttendances(opts)
	if err != nil {
		return nil, 0, err
	}
	return attendances, total, nil
}

// GetAllFilteredAttendancesByBatchSlug retrieves all attendances with pagination and filtering options
func (s *AttendanceServices) GetAllFilteredAttendancesByBatchSlug(batchSlug string, opts utils.QueryOptions) ([]models.Attendance, int64, error) {
	attendances, total, err := s.attendanceRepo.GetAllFilteredAttendancesByBatchSlug(batchSlug, opts)
	if err != nil {
		return nil, 0, err
	}
	return attendances, total, nil
}

// GetAttendanceByID retrieves a single attendance by its ID
func (s *AttendanceServices) GetAttendanceByID(attendanceID uuid.UUID) (*models.Attendance, error) {
	return s.attendanceRepo.FindByID(attendanceID)
}

// BulkUpsertAttendance for bulk attendance services
func (s *AttendanceServices) BulkUpsertAttendance(user *utils.Claims, batchID uuid.UUID, req *dto.BulkAttendanceRequest) ([]models.Attendance, error) {
	var results []models.Attendance

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		meetings, err := s.meetingRepo.WithTx(tx).GetMeetingsByBatchID(batchID)
		if err != nil {
			return fmt.Errorf("failed to fetch meetings for batch: %w", err)
		}

		validMeetings := make(map[uuid.UUID]bool)
		for _, m := range meetings {
			validMeetings[m.ID] = true
		}

		for _, item := range req.Attendances {
			if !validMeetings[item.MeetingID] {
				return fmt.Errorf("Invalid meeting ID %s for batch", item.MeetingID)
			}

			hasPaid, err := s.purchaseRepo.WithTx(tx).HasPaid(item.UserID, batchID)
			if err != nil {
				return fmt.Errorf("Failed to check payment for user %s: %w", item.UserID, err)
			}
			if !hasPaid {
				return fmt.Errorf("User %s has not purchased the batch", item.UserID)
			}

			existing, err := s.attendanceRepo.WithTx(tx).GetByMeetingAndUser(item.MeetingID, item.UserID)
			if err == nil && existing != nil {
				existing.IsPresent = item.IsPresent
				existing.Note = item.Note
				existing.UpdatedBy = user.UserID

				if err := s.attendanceRepo.WithTx(tx).UpdateByMeetingAndUser(item.MeetingID, item.UserID, existing); err != nil {
					return err
				}
				results = append(results, *existing)
			} else {
				newAttendance := &models.Attendance{
					MeetingID: item.MeetingID,
					UserID:    item.UserID,
					IsPresent: item.IsPresent,
					Note:      item.Note,
					UpdatedBy: user.UserID,
				}
				if err := s.attendanceRepo.WithTx(tx).Create(newAttendance); err != nil {
					return err
				}
				results = append(results, *newAttendance)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}
