package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
)

// ITestimonialService interface
type ITestimonialService interface {
	GetAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.Testimonial, int64, error)
	GetByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Testimonial, int64, error)
	Create(ctx context.Context, req *dto.CreateTestimonialRequest, batchID, userID uuid.UUID) (*models.Testimonial, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTestimonialRequest, userID uuid.UUID) (*models.Testimonial, error)
	Delete(ctx context.Context, id, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Testimonial, error)
}

// TestimonialService service
type TestimonialService struct {
	testimonialRepo repository.ITestimonialRepository
	purchaseService IPurchaseService
	batchRepo       repository.IBatchRepository
}

// NewTestimonialService init service
func NewTestimonialService(testimonialRepo repository.ITestimonialRepository, purchaseService IPurchaseService, batchRepo repository.IBatchRepository) ITestimonialService {
	return &TestimonialService{testimonialRepo: testimonialRepo, purchaseService: purchaseService, batchRepo: batchRepo}
}

// GetAllFiltered Get with filter & pagination
func (s *TestimonialService) GetAllFiltered(ctx context.Context, opts utils.QueryOptions) ([]models.Testimonial, int64, error) {
	return s.testimonialRepo.GetAllFiltered(ctx, opts)
}

// GetByBatchSlugFiltered Get with filter & pagination
func (s *TestimonialService) GetByBatchSlugFiltered(ctx context.Context, batchSlug string, opts utils.QueryOptions) ([]models.Testimonial, int64, error) {
	return s.testimonialRepo.GetByBatchSlugFiltered(ctx, batchSlug, opts)
}

// GetByID get by id service
func (s *TestimonialService) GetByID(ctx context.Context, id uuid.UUID) (*models.Testimonial, error) {
	return s.testimonialRepo.GetByID(ctx, id)
}

// Create services
func (s *TestimonialService) Create(ctx context.Context, req *dto.CreateTestimonialRequest, batchID, userID uuid.UUID) (*models.Testimonial, error) {
	_, err := s.batchRepo.FindByID(ctx, batchID)
	if err != nil {
		return nil, err
	}

	existing, err := s.testimonialRepo.GetByUserAndBatch(ctx, userID, batchID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("testimonial already exists")
	}

	isPaid, err := s.purchaseService.HasPaid(ctx, userID, batchID)
	if err != nil {
		return nil, err
	}
	if !isPaid {
		return nil, fmt.Errorf("user has not paid for this batch")
	}

	testimonial := &models.Testimonial{
		UserID:      userID,
		BatchID:     batchID,
		Rating:      req.Rating,
		Title:       req.Title,
		Description: req.Description,
	}
	if err := s.testimonialRepo.Create(ctx, testimonial); err != nil {
		return nil, err
	}
	if testimonial, err = s.testimonialRepo.GetByID(ctx, testimonial.ID); err != nil {
		return nil, err
	}

	return testimonial, nil
}

// Update update
func (s *TestimonialService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateTestimonialRequest, userID uuid.UUID) (*models.Testimonial, error) {
	existing, err := s.testimonialRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}

	isPaid, err := s.purchaseService.HasPaid(ctx, userID, existing.BatchID)
	if err != nil {
		return nil, err
	}
	if !isPaid {
		return nil, fmt.Errorf("user has not paid for this batch")
	}

	// Copy field yang tidak nil saja
	if err := copier.CopyWithOption(&existing, req, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return nil, err
	}

	if err := s.testimonialRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete delete
func (s *TestimonialService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	existing, err := s.testimonialRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return fmt.Errorf("forbidden")
	}

	isPaid, err := s.purchaseService.HasPaid(ctx, userID, existing.BatchID)
	if err != nil {
		return err
	}
	if !isPaid {
		return fmt.Errorf("user has not paid for this batch")
	}

	return s.testimonialRepo.Delete(ctx, id)
}
