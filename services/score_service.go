package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IScoreService interface
type IScoreService interface {
	GetScoresByBatchUser(ctx context.Context, batchID uuid.UUID, user *utils.Claims) (*dto.ScoreResponse, error)
	GetScoresByBatchStudentSlug(ctx context.Context, batchSlug, studentIDParam string) (*dto.ScoreResponse, error)
}

// ScoreService provides methods for managing scores
type ScoreService struct {
	batchRepo       repository.IBatchRepository
	meetingRepo     repository.IMeetingRepository
	submissionRepo  repository.ISubmisssionRepository
	quizRepo        repository.IQuizRepository
	purchaseService IPurchaseService
	db              *gorm.DB
}

// NewScoreService creates a new instance of ScoreService
func NewScoreService(db *gorm.DB, batchRepo repository.IBatchRepository, meetingRepo repository.IMeetingRepository,
	purchaseService IPurchaseService, quizRepo repository.IQuizRepository, submissionRepo repository.ISubmisssionRepository) IScoreService {
	return &ScoreService{db: db, batchRepo: batchRepo, meetingRepo: meetingRepo, purchaseService: purchaseService,
		quizRepo: quizRepo, submissionRepo: submissionRepo}
}

func (s *ScoreService) checkUserAccess(ctx context.Context, user *utils.Claims, batchID uuid.UUID) (bool, error) {
	// Cari batch info dari batchID
	batch, err := s.batchRepo.FindByID(ctx, batchID) // balikin batchSlug & batchID
	if err != nil {
		return false, err
	}

	// Kalau role teacher, cek apakah dia mengajar batch ini
	if user.Role == string(models.RoleTypeGuru) {
		return s.meetingRepo.IsBatchOwnedByUser(ctx, user.UserID, batch.Slug)
	}

	// Kalau student, cek pembayaran
	if user.Role == string(models.RoleTypeSiswa) {
		return s.purchaseService.HasPaid(ctx, user.UserID, batch.ID)
	}

	// Role lain tidak diizinkan
	return false, nil
}

// GetScoresByBatchUser retrieves assignment & quiz scores of a user in a batch
func (s *ScoreService) GetScoresByBatchUser(ctx context.Context, batchID uuid.UUID, user *utils.Claims) (*dto.ScoreResponse, error) {
	allowed, err := s.checkUserAccess(ctx, user, batchID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("forbidden")
	}

	// Ambil assignment scores
	assignScores, err := s.submissionRepo.GetAssignmentsWithScoresByBatchUser(ctx, batchID, user.UserID)
	if err != nil {
		return nil, err
	}

	// Ambil quiz scores
	quizScores, err := s.quizRepo.GetQuizzesWithScoresByBatchUser(ctx, batchID, user.UserID)
	if err != nil {
		return nil, err
	}

	return &dto.ScoreResponse{
		Assignments: assignScores,
		Quizzes:     quizScores,
	}, nil
}

// GetScoresByBatchStudentSlug resolves batchSlug + studentID, then retrieves scores
func (s *ScoreService) GetScoresByBatchStudentSlug(ctx context.Context, batchSlug, studentIDParam string) (*dto.ScoreResponse, error) {
	if batchSlug == "" {
		return nil, fmt.Errorf("batchSlug is required")
	}
	if studentIDParam == "" {
		return nil, fmt.Errorf("studentID is required")
	}

	studentID, err := uuid.Parse(studentIDParam)
	if err != nil {
		return nil, fmt.Errorf("invalid studentID: %w", err)
	}

	// Cari batchID dari slug
	batch, err := s.batchRepo.GetBatchBySlug(ctx, batchSlug)
	if err != nil {
		return nil, err
	}

	// Ambil assignment scores
	assignScores, err := s.submissionRepo.GetAssignmentsWithScoresByBatchUser(ctx, batch.ID, studentID)
	if err != nil {
		return nil, err
	}

	// Ambil quiz scores
	quizScores, err := s.quizRepo.GetQuizzesWithScoresByBatchUser(ctx, batch.ID, studentID)
	if err != nil {
		return nil, err
	}

	return &dto.ScoreResponse{
		Assignments: assignScores,
		Quizzes:     quizScores,
	}, nil
}
