package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// IBatchService interface
type IBatchService interface {
	GetAllFilteredBatches(ctx context.Context, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetBatchBySlug(ctx context.Context, slug string) (*models.Batch, error)
	CreateBatch(ctx context.Context, courseID uuid.UUID, body *dto.CreateBatchRequest) (*models.Batch, error)
	UpdateBatch(ctx context.Context, id uuid.UUID, body *dto.UpdateBatchRequest) (*models.Batch, error)
	DeleteBatch(ctx context.Context, batchID uuid.UUID) error
	GetBatchByCourseSlug(ctx context.Context, courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetBatchesPurchasedByUser(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
	GetBatchesTaughtByGuru(ctx context.Context, guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error)
	CalculateProgress(ctx context.Context, batchID, userID uuid.UUID) (float64, error)
}

// BatchService provides methods for managing batches
type BatchService struct {
	repo           repository.IBatchRepository
	userRepo       repository.IUserRepository
	courseRepo     repository.ICourseRepository
	assignmentRepo repository.IAssignmentRepository
	submissionRepo repository.ISubmisssionRepository
	db             *gorm.DB
	fileService    IFileService
}

// NewBatchService creates a new instance of BatchService
func NewBatchService(repo repository.IBatchRepository, userRepo repository.IUserRepository, courseRepo repository.ICourseRepository,
	assignmentRepo repository.IAssignmentRepository,
	submissionRepo repository.ISubmisssionRepository, db *gorm.DB, fileService IFileService) IBatchService {
	return &BatchService{repo: repo, userRepo: userRepo, courseRepo: courseRepo, assignmentRepo: assignmentRepo, submissionRepo: submissionRepo, db: db, fileService: fileService}
}

// GetAllFilteredBatches retrieves all batches with pagination and filtering options
func (s *BatchService) GetAllFilteredBatches(ctx context.Context, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	batches, total, err := s.repo.GetAllFilteredBatches(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

// GetBatchBySlug retrieves a batch by its slug
func (s *BatchService) GetBatchBySlug(ctx context.Context, slug string) (*models.Batch, error) {
	batch, err := s.repo.GetBatchBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return batch, nil
}

// CreateBatch creates a new batch with the provided details
func (s *BatchService) CreateBatch(ctx context.Context, courseID uuid.UUID, body *dto.CreateBatchRequest) (*models.Batch, error) {
	var batch models.Batch

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Validasi course ID
		_, err := s.courseRepo.WithTx(tx).FindByID(ctx, courseID)
		if err != nil {
			return err
		}

		// Copy data dari body ke batch
		copier.Copy(&batch, body)

		slug := utils.GenerateUniqueSlug(ctx, body.Title, s.repo)

		// Parse waktu dari string ke time.Time
		parsedStart, err := time.Parse("15:04", body.StartTime)
		if err != nil {
			return err
		}
		parsedEnd, err := time.Parse("15:04", body.EndTime)
		if err != nil {
			return err
		}

		batch.Slug = slug
		batch.CourseID = courseID
		batch.StartTime = parsedStart.Format("15:04")
		batch.EndTime = parsedEnd.Format("15:04")

		// Simpan batch utama
		if err := s.repo.WithTx(tx).Create(ctx, &batch); err != nil {
			return err
		}

		// Simpan BatchDays
		for _, day := range body.Days {
			batchDay := models.BatchDay{
				BatchID: batch.ID,
				Day:     day,
			}
			if err := tx.WithContext(ctx).Create(&batchDay).Error; err != nil {
				return err
			}
		}

		// 🔥 Simpan BatchGroups
		for _, groupType := range body.GroupTypes {
			bg := models.BatchGroup{
				BatchID:   batch.ID,
				GroupType: groupType,
			}
			if err := tx.WithContext(ctx).Create(&bg).Error; err != nil {
				return err
			}
		}

		// Ambil data batch lengkap setelah insert
		updated, err := s.repo.WithTx(tx).FindByID(ctx, batch.ID)
		if err != nil {
			return fmt.Errorf("gagal mengambil batch setelah dibuat: %w", err)
		}

		batch = utils.Safe(updated, models.Batch{})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// UpdateBatch updates an existing batch with the provided details
func (s *BatchService) UpdateBatch(ctx context.Context, id uuid.UUID, body *dto.UpdateBatchRequest) (*models.Batch, error) {
	var batch models.Batch

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		// Ambil batch dari database
		batchPtr, err := s.repo.WithTx(tx).FindByID(ctx, id)
		if err != nil {
			return err
		}
		batch = utils.Safe(batchPtr, models.Batch{})

		// Copy field yang tidak nil dari request ke model
		if err := copier.CopyWithOption(&batch, body, copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		}); err != nil {
			return err
		}

		// Parse StartTime dan EndTime
		if body.StartTime != nil {
			parsedStart, err := time.Parse("15:04", *body.StartTime)
			if err != nil {
				return fmt.Errorf("invalid start_time: %w", err)
			}
			batch.StartTime = parsedStart.Format("15:04")
		}

		if body.EndTime != nil {
			parsedEnd, err := time.Parse("15:04", *body.EndTime)
			if err != nil {
				return fmt.Errorf("invalid end_time: %w", err)
			}
			batch.EndTime = parsedEnd.Format("15:04")
		}

		// Simpan perubahan batch
		if err := s.repo.WithTx(tx).Update(ctx, &batch); err != nil {
			return err
		}

		// Update BatchDays jika dikirim
		if body.Days != nil {
			if err := tx.WithContext(ctx).Where("batch_id = ?", batch.ID).Delete(&models.BatchDay{}).Error; err != nil {
				return err
			}
			for _, day := range *body.Days {
				batchDay := models.BatchDay{
					BatchID: batch.ID,
					Day:     day,
				}
				if err := tx.WithContext(ctx).Create(&batchDay).Error; err != nil {
					return err
				}
			}
		}

		// ✅ Update BatchGroups jika GroupTypes dikirim
		if body.GroupTypes != nil {
			if err := tx.WithContext(ctx).Where("batch_id = ?", batch.ID).Delete(&models.BatchGroup{}).Error; err != nil {
				return err
			}
			for _, gtype := range *body.GroupTypes {
				batchGroup := models.BatchGroup{
					BatchID:   batch.ID,
					GroupType: gtype,
				}
				if err := tx.WithContext(ctx).Create(&batchGroup).Error; err != nil {
					return err
				}
			}
		}

		// Ambil data terbaru
		updated, err := s.repo.WithTx(tx).FindByID(ctx, batch.ID)
		if err != nil {
			return fmt.Errorf("gagal mengambil batch setelah diupdate: %w", err)
		}

		batch = *updated
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// DeleteBatch deletes a batch by its ID
func (s *BatchService) DeleteBatch(ctx context.Context, batchID uuid.UUID) error {
	var batch models.Batch
	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		var err error
		batchRsp, err := s.repo.WithTx(tx).FindByID(ctx, batchID)
		if err != nil {
			return err
		}

		if batchRsp == nil {
			return fmt.Errorf("batch tidak ditemukan")
		}

		batch = utils.Safe(batchRsp, models.Batch{})

		// Hapus batch (images akan ikut terhapus karena cascade)
		if err := s.repo.WithTx(tx).DeleteByID(ctx, batchID); err != nil {
			return err
		}

		return nil
	})

	// Hapus file gambar

	if err := s.fileService.DeleteFile(batch.BatchThumbnail); err != nil {
		log.Errorf("Gagal hapus file %s: %v", batch.BatchThumbnail, err)
	}

	if err != nil {
		return err
	}
	return nil

}

// GetBatchByCourseSlug is function for get all batches by course slug
func (s *BatchService) GetBatchByCourseSlug(ctx context.Context, courseID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	batches, total, err := s.repo.GetAllFilteredBatchesByCourseSlug(ctx, courseID, opts)
	if err != nil {
		return nil, 0, err
	}
	return batches, total, nil
}

// GetBatchesPurchasedByUser is service for get batches where the user has purchased
func (s *BatchService) GetBatchesPurchasedByUser(ctx context.Context, userID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	return s.repo.GetBatchesByUserPurchaseFiltered(ctx, userID, opts)
}

// GetBatchesTaughtByGuru is service for get batches where teacher was taughted
func (s *BatchService) GetBatchesTaughtByGuru(ctx context.Context, guruID uuid.UUID, opts utils.QueryOptions) ([]models.Batch, int64, error) {
	return s.repo.GetBatchesByGuruMeetingRelationFiltered(ctx, guruID, opts)
}

// CalculateProgress service calculate progress
func (s *BatchService) CalculateProgress(ctx context.Context, batchID, userID uuid.UUID) (float64, error) {
	// Hitung total assignment di batch
	totalAssignments, err := s.assignmentRepo.CountByBatchID(ctx, batchID)
	if err != nil {
		return 0, err
	}
	if totalAssignments == 0 {
		return 0, nil // biar ga bagi nol
	}

	// Hitung total submission user yg sudah dikumpulkan di batch tsb
	completedSubmissions, err := s.submissionRepo.CountCompletedByBatchUser(ctx, batchID, userID)
	if err != nil {
		return 0, err
	}

	// Hitung persentase progress
	progress := (float64(completedSubmissions) / float64(totalAssignments)) * 100
	return progress, nil
}
