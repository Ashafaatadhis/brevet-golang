package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// MaterialService provides methods for managing materials
type MaterialService struct {
	materialRepo *repository.MaterialRepository
	meetingRepo  *repository.MeetingRepository
	purchaseRepo *repository.PurchaseRepository
	fileService  *FileService
	db           *gorm.DB
}

// NewMaterialService creates a new instance of MaterialService
func NewMaterialService(materialRepo *repository.MaterialRepository, meetingRepository *repository.MeetingRepository, purchaseRepo *repository.PurchaseRepository, fileService *FileService, db *gorm.DB) *MaterialService {
	return &MaterialService{materialRepo: materialRepo, meetingRepo: meetingRepository, purchaseRepo: purchaseRepo, fileService: fileService, db: db}
}

// GetAllFilteredMaterial retrieves all materials with pagination and filtering options
func (s *MaterialService) GetAllFilteredMaterial(opts utils.QueryOptions) ([]models.Material, int64, error) {
	materials, total, err := s.materialRepo.GetAllFilteredMaterial(opts)
	if err != nil {
		return nil, 0, err
	}
	return materials, total, nil
}

// GetAllFilteredMaterialsByMeetingID retrieves all materials with pagination and filtering options
func (s *MaterialService) GetAllFilteredMaterialsByMeetingID(meetingID uuid.UUID, user *utils.Claims, opts utils.QueryOptions) ([]models.Material, int64, error) {
	if user.Role == string(models.RoleTypeGuru) {
		ok, err := s.meetingRepo.IsMeetingTaughtByUser(meetingID, user.UserID)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, err
		}
	}

	materials, total, err := s.materialRepo.GetAllFilteredMaterialsByMeetingID(meetingID, opts)
	if err != nil {
		return nil, 0, err
	}
	return materials, total, nil
}

// GetMaterialByID retrieves a single materials by its ID
func (s *MaterialService) GetMaterialByID(user *utils.Claims, materialID uuid.UUID) (*models.Material, error) {
	material, err := s.materialRepo.FindByID(materialID)
	if err != nil {
		return nil, err
	}

	switch user.Role {
	case string(models.RoleTypeAdmin):
		// ✅ Admin bebas ambil
		return material, nil

	case string(models.RoleTypeGuru):
		// 🔒 Guru hanya boleh jika ngajar di meeting terkait
		isGuru, err := s.meetingRepo.IsUserTeachingInMeeting(user.UserID, material.MeetingID)
		if err != nil {
			return nil, err
		}
		if !isGuru {
			return nil, fmt.Errorf("Anda bukan pengajar di meeting ini")
		}
		return material, nil

	case string(models.RoleTypeSiswa):
		meeting, err := s.meetingRepo.FindByID(material.MeetingID)
		if err != nil {
			return nil, err
		}
		// 🔒 Siswa hanya bisa jika sudah beli batch meeting tersebut
		isPurchased, err := s.purchaseRepo.HasPaid(user.UserID, meeting.BatchID)
		if err != nil {
			return nil, err
		}
		if !isPurchased {
			return nil, fiber.NewError(fiber.StatusForbidden, "Anda belum terdaftar di batch meeting ini")
		}
		return material, nil

	default:
		return nil, fiber.NewError(fiber.StatusForbidden, "Role tidak dikenali")
	}
}

// CreateMaterial creates a new material with the provided details
func (s *MaterialService) CreateMaterial(user *utils.Claims, meetingID uuid.UUID, body *dto.CreateMaterialRequest) (*models.Material, error) {
	var material models.Material

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {

		meeting, err := s.meetingRepo.WithTx(tx).FindByID(meetingID)
		if err != nil {
			return err
		}

		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(meeting.ID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, meeting.ID)
			}
		}

		materialPtr := &models.Material{
			ID:          uuid.New(),
			MeetingID:   meetingID,
			URL:         body.URL,
			Title:       body.Title,
			Description: utils.SafeNil(body.Description),
		}

		if err := s.materialRepo.WithTx(tx).Create(materialPtr); err != nil {
			return err
		}

		// ✅ Ambil ulang dari DB untuk dapet semua kolom yang terisi otomatis (CreatedAt, dll)
		updated, err := s.materialRepo.WithTx(tx).FindByID(materialPtr.ID)
		if err != nil {
			return err
		}
		material = utils.Safe(updated, models.Material{})

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &material, nil
}

// UpdateMaterial updates an existing material and its files
func (s *MaterialService) UpdateMaterial(user *utils.Claims, materialID uuid.UUID, body *dto.UpdateMaterialRequest) (*models.Material, error) {
	var updatedMaterial models.Material

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		material, err := s.materialRepo.WithTx(tx).FindByID(materialID)
		if err != nil {
			return err
		}

		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(material.MeetingID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, material.MeetingID)
			}
		}

		// Copy field yang tidak nil saja
		if err := copier.CopyWithOption(&material, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		if err := s.materialRepo.WithTx(tx).Update(material); err != nil {
			return err
		}

		// Ambil ulang assignment lengkap
		fresh, err := s.materialRepo.WithTx(tx).FindByID(material.ID)
		if err != nil {
			return err
		}
		updatedMaterial = utils.Safe(fresh, models.Material{})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &updatedMaterial, nil
}

// DeleteMaterial deletes an material and
func (s *MaterialService) DeleteMaterial(user *utils.Claims, materialID uuid.UUID) error {
	var material models.Material

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		materialRsp, err := s.materialRepo.WithTx(tx).FindByID(materialID)
		if err != nil {
			return err
		}

		if user.Role == string(models.RoleTypeGuru) {
			ok, err := s.meetingRepo.IsMeetingTaughtByUser(materialRsp.MeetingID, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to check meeting-teacher relation: %w", err)
			}
			if !ok {
				return fmt.Errorf("forbidden: user %s is not assigned to teach meeting %s", user.UserID, materialRsp.MeetingID)
			}
		}

		material = utils.Safe(materialRsp, models.Material{})

		// Hapus dari DB (files ikut kehapus karena CASCADE)
		if err := s.materialRepo.WithTx(tx).DeleteByID(materialID); err != nil {
			return err
		}

		return nil
	})

	// Setelah commit, hapus file dari cloud atau disk
	if err := s.fileService.DeleteFile(material.URL); err != nil {
		log.Errorf("Gagal hapus file %s: %v", material.URL, err)
	}

	if err != nil {
		return err
	}
	return nil
}
