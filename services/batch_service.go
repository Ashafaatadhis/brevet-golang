package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/policies"
	"brevet-api/repository"
	"brevet-api/utils"
	"fmt"

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

	fmt.Printf("courseRepo is nil? %v\n", s.courseRepo == nil)
	fmt.Printf("courseRepo pointer: %p\n", s.courseRepo.FindByID)

	a, err := s.courseRepo.FindByID(tx, courseID)
	if err != nil {
		return nil, err
	}

	fmt.Print(a, "TAI\n")

	var batch models.Batch
	copier.Copy(&batch, body)

	slug := utils.GenerateUniqueSlug(body.Title, s.repo)

	batch.Slug = slug
	batch.CourseID = courseID

	if err := s.repo.CreateTx(tx, &batch); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &batch, nil
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
	// 	course.Slug = slug
	// }

	if err := s.repo.UpdateTx(tx, batch); err != nil {
		return nil, err
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

// THIS IN BOTTOM SERVICES IS FOR ASSIGN TEACHER TO BATCH

// AddTeacherToBatch adds a teacher to a batch
func (s *BatchService) AddTeacherToBatch(batchID uuid.UUID, body *dto.CreateBatchTeacherRequest) (*models.User, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	_, err := s.repo.FindByIDTx(tx, batchID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByIDTx(tx, body.UserID)
	if err != nil {
		return nil, err
	}

	if !policies.CanBeAssignedAsTeacher(user) {
		return nil, fmt.Errorf("user bukan teacher")
	}

	// Cek apakah user sudah jadi teacher di batch ini
	exists, err := s.repo.IsTeacherAssigned(batchID, body.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("user sudah jadi teacher di batch ini")
	}

	var batch models.BatchTeacher
	if err := copier.Copy(&batch, body); err != nil {
		return nil, err
	}

	batch.BatchID = batchID

	if err := s.repo.CreateBatchTeacherTx(tx, &batch); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetTeacherInBatch get teacher in batch by batch id
func (s *BatchService) GetTeacherInBatch(batchID uuid.UUID, opts utils.QueryOptions) ([]models.User, int64, error) {
	if _, err := s.repo.FindByID(batchID); err != nil {
		return nil, 0, err
	}

	users, total, err := s.repo.GetAllTeacherInBatch(batchID, opts)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// DeleteTeacherFromBatch deletes a teacher from a batch
func (s *BatchService) DeleteTeacherFromBatch(batchID uuid.UUID, userID uuid.UUID) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Optional: cek dulu apakah ada batch-nya
	_, err := s.repo.FindBatchTeacherByBatchIDAndUserIDTx(tx, batchID, userID)
	if err != nil {
		return err
	}

	// Hapus batch (images akan ikut terhapus karena cascade)
	if err := s.repo.DeleteTeacherByIDTx(tx, batchID, userID); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
