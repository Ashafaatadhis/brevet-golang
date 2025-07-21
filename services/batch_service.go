package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// BatchService provides methods for managing batches
type BatchService struct {
	repo        *repository.BatchRepository
	userRepo    *repository.UserRepository
	courseRepo  *repository.CourseRepository
	db          *gorm.DB
	fileService *FileService
}

// NewBatchService creates a new instance of BatchService
func NewBatchService(repo *repository.BatchRepository, userRepo *repository.UserRepository, courseRepo *repository.CourseRepository, db *gorm.DB, fileService *FileService) *BatchService {
	return &BatchService{repo: repo, userRepo: userRepo, courseRepo: courseRepo, db: db, fileService: fileService}
}

// GetAllFilteredBatches retrieves all batches with pagination and filtering options
func (s *BatchService) GetAllFilteredBatches(opts utils.QueryOptions) ([]models.Batch, int64, error) {
	batches, total, err := s.repo.GetAllFilteredBatches(opts)
	if err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

// GetBatchBySlug retrieves a batch by its slug
func (s *BatchService) GetBatchBySlug(slug string) (*models.Batch, error) {
	batch, err := s.repo.GetBatchBySlug(slug)
	if err != nil {
		return nil, err
	}
	return batch, nil
}

// CreateBatch creates a new batch with the provided details
func (s *BatchService) CreateBatch(courseID uuid.UUID, body *dto.CreateBatchRequest) (*models.Batch, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Validasi course ID
	_, err := s.courseRepo.FindByID(tx, courseID)
	if err != nil {
		return nil, err
	}

	var batch models.Batch
	copier.Copy(&batch, body)

	slug := utils.GenerateUniqueSlug(body.Title, s.repo)

	// Parse waktu dari string ke time.Time
	parsedStart, err := time.Parse("15:04", body.StartTime)
	if err != nil {
		return nil, err
	}
	parsedEnd, err := time.Parse("15:04", body.EndTime)
	if err != nil {
		return nil, err
	}

	batch.Slug = slug
	batch.CourseID = courseID
	batch.StartTime = parsedStart
	batch.EndTime = parsedEnd

	// Simpan batch utama
	if err := s.repo.CreateTx(tx, &batch); err != nil {
		return nil, err
	}

	// Simpan BatchDays
	for _, day := range body.Days {
		batchDay := models.BatchDay{
			BatchID: batch.ID,
			Day:     day,
		}

		if err := tx.Create(&batchDay).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.repo.FindByID(batch.ID)
}

// UpdateBatch updates an existing batch with the provided details
func (s *BatchService) UpdateBatch(id uuid.UUID, body *dto.UpdateBatchRequest) (*models.Batch, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	batch, err := s.repo.FindByIDTx(tx, id)
	if err != nil {
		return nil, err
	}

	// Copy field yang tidak nil saja
	if err := copier.CopyWithOption(batch, body, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return nil, err
	}

	// Optional: regenerate slug kalau Title berubah
	// if body.Title != nil {
	// 	slug := utils.GenerateUniqueSlug(*body.Title, s.repo)
	// 	batch.Slug = slug
	// }

	// Parse waktu dari string ke time.Time
	if body.StartTime != nil {
		parsedStart, err := time.Parse("15:04", *body.StartTime)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time: %w", err)
		}
		batch.StartTime = parsedStart
	}

	if body.EndTime != nil {
		parsedEnd, err := time.Parse("15:04", *body.EndTime)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time: %w", err)
		}
		batch.EndTime = parsedEnd
	}

	if err := s.repo.UpdateTx(tx, batch); err != nil {
		return nil, err
	}

	// Update BatchDays jika diberikan
	if body.Days != nil {
		if err := tx.Where("batch_id = ?", batch.ID).Delete(&models.BatchDay{}).Error; err != nil {
			return nil, err
		}
		for _, day := range *body.Days {
			batchDay := models.BatchDay{
				BatchID: batch.ID,
				Day:     day,
			}
			if err := tx.Create(&batchDay).Error; err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.repo.FindByID(batch.ID)
}

// DeleteBatch deletes a batch by its ID
func (s *BatchService) DeleteBatch(batchID uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Optional: cek dulu apakah ada batch-nya
	batch, err := s.repo.FindByIDTx(tx, batchID)
	if err != nil {
		return err
	}

	// Hapus batch (images akan ikut terhapus karena cascade)
	if err := s.repo.DeleteByIDTx(tx, batchID); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Hapus file gambar
	if err := s.fileService.DeleteFile(batch.BatchThumbnail); err != nil {
		// Log error tapi lanjut
		log.Errorf("Gagal hapus file %s: %v", batch.BatchThumbnail, err)
	}

	return nil
}

// GetBatchByCourseSlug is function for get all batches by course slug
func (s *BatchService) GetBatchByCourseSlug(courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	batches, total, err := s.repo.GetAllFilteredBatchesByCourseSlug(courseID, opts)
	if err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

// GetBatchesPurchasedByUser is service for get batches where the user has purchased
func (s *BatchService) GetBatchesPurchasedByUser(userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	return s.repo.GetBatchesByUserPurchaseFiltered(userID, opts)
}

// GetBatchesTaughtByGuru is service for get batches where teacher was taughted
func (s *BatchService) GetBatchesTaughtByGuru(guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	return s.repo.GetBatchesByGuruMeetingRelationFiltered(guruID, opts)
}
