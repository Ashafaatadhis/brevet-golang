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

// GetAttendanceByID retrieves a single attendance by its ID
func (s *AttendanceServices) GetAttendanceByID(attendanceID uuid.UUID) (*models.Attendance, error) {
	return s.attendanceRepo.FindByID(attendanceID)
}

// BulkUpsertAttendance for bulk attendance services
func (s *AttendanceServices) BulkUpsertAttendance(user *utils.Claims, meetingID uuid.UUID, req *dto.BulkAttendanceRequest) ([]models.Attendance, error) {
	var results []models.Attendance

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		meeting, err := s.meetingRepo.WithTx(tx).FindByID(meetingID)
		if err != nil {
			return fmt.Errorf("Meeting not found")
		}

		for _, item := range req.Attendances {
			hasPaid, err := s.purchaseRepo.WithTx(tx).HasPaid(item.UserID, meeting.BatchID)
			if err != nil {
				return fmt.Errorf("Failed to check payment for user %s: %w", item.UserID, err)
			}

			if !hasPaid {
				return fmt.Errorf("User %s has not purchased the batch", item.UserID)
			}

			existing, err := s.attendanceRepo.WithTx(tx).GetByMeetingAndUser(meetingID, item.UserID)

			if err == nil && existing != nil {
				// Update
				existing.Status = item.Status
				existing.Note = item.Note
				existing.UpdatedBy = user.UserID

				if err := s.attendanceRepo.WithTx(tx).UpdateByMeetingAndUser(meetingID, item.UserID, existing); err != nil {
					return err
				}
				results = append(results, *existing)
			} else {
				// Create
				newAttendance := &models.Attendance{
					MeetingID: meetingID,
					UserID:    item.UserID,
					Status:    item.Status,
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
