package services

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/repository"
	"brevet-api/utils"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurchaseService provides methods for managing purchases
type PurchaseService struct {
	purchaseRepo *repository.PurchaseRepository
	userRepo     *repository.UserRepository
	db           *gorm.DB
}

// NewPurchaseService creates a new instance of PurchaseService
func NewPurchaseService(purchaseRepository *repository.PurchaseRepository, db *gorm.DB) *PurchaseService {
	return &PurchaseService{purchaseRepo: purchaseRepository, db: db}
}

// GetAllFilteredPurchases retrieves all purchases with pagination and filtering options
func (s *PurchaseService) GetAllFilteredPurchases(opts utils.QueryOptions) ([]models.Purchase, int64, error) {
	purchases, total, err := s.purchaseRepo.GetAllFilteredPurchases(opts)
	if err != nil {
		return nil, 0, err
	}
	return purchases, total, nil
}

// GetPurchaseByID retrieves a course by its slug
func (s *PurchaseService) GetPurchaseByID(id uuid.UUID) (*models.Purchase, error) {
	purchase, err := s.purchaseRepo.GetPurchaseByID(id)
	if err != nil {
		return nil, err
	}
	return purchase, nil
}

// HasPaid is for check user has paid or not
func (s *PurchaseService) HasPaid(userID uuid.UUID, batchID uuid.UUID) (bool, error) {
	return s.purchaseRepo.HasPaid(userID, batchID)
}

// GetPaidBatchIDs for get all batch where user has paid
func (s *PurchaseService) GetPaidBatchIDs(userID string) ([]string, error) {
	return s.purchaseRepo.GetPaidBatchIDs(userID)
}

// CreatePurchase is for create purchase
func (s *PurchaseService) CreatePurchase(userID uuid.UUID, batchID uuid.UUID) (*models.Purchase, error) {
	tx := s.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()
	// 1. Cek apakah sudah pernah beli (pakai tx)
	hasPaid, err := s.purchaseRepo.WithTx(tx).HasPurchaseWithStatus(userID, batchID,
		[]models.PaymentStatus{models.Pending, models.WaitingConfirmation, models.Paid}...)
	if err != nil {
		return nil, err
	}
	if hasPaid {
		return nil, errors.New("Anda sudah memiliki transaksi untuk batch ini")
	}

	// 2. Ambil user
	var user *models.User
	user, err = s.userRepo.FindByIDTx(tx, userID)
	if err != nil {
		return nil, fmt.Errorf("User tidak ditemukan: %w", err)
	}

	if user.Profile == nil || user.Profile.GroupType == nil {
		return nil, fmt.Errorf("User belum memiliki GroupType yang valid")
	}

	// 3. Ambil harga
	price, err := s.purchaseRepo.WithTx(tx).GetPriceByGroupType(user.Profile.GroupType)
	if err != nil {
		return nil, fmt.Errorf("harga untuk group_type '%s' tidak ditemukan: %w", *user.Profile.GroupType, err)
	}

	expiredAt := time.Now().Add(24 * time.Hour)
	// 4. Buat record purchase
	purchase := &models.Purchase{
		UserID:        &userID,
		BatchID:       &batchID,
		PriceID:       price.ID,
		ExpiredAt:     &expiredAt,
		PaymentStatus: models.Pending,
	}

	if err := s.purchaseRepo.WithTx(tx).Create(purchase); err != nil {
		return nil, err
	}

	// 5. Commit jika semua berhasil
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	purchaseWithPrice, err := s.purchaseRepo.GetPurchaseByID(purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil ulang purchase: %w", err)
	}

	return purchaseWithPrice, nil

}

// UpdateStatusPayment verification payment service
func (s *PurchaseService) UpdateStatusPayment(purchaseID uuid.UUID, body *dto.UpdateStatusPayment) (*models.Purchase, error) {
	purchase, err := s.purchaseRepo.GetPurchaseByID(purchaseID)
	if err != nil {
		return nil, fmt.Errorf("data tidak ditemukan: %w", err)
	}

	if purchase.PaymentStatus != models.WaitingConfirmation {
		return nil, fmt.Errorf("status pembayaran tidak bisa diverifikasi")
	}

	purchase.PaymentStatus = body.PaymentStatus

	err = s.purchaseRepo.Update(purchase)
	if err != nil {
		return nil, err
	}

	purchaseWithPrice, err := s.purchaseRepo.GetPurchaseByID(purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil ulang purchase: %w", err)
	}

	return purchaseWithPrice, nil
}

// PayPurchase is for pay purchase
func (s *PurchaseService) PayPurchase(userID uuid.UUID, purchaseID uuid.UUID, proofURL string) (*models.Purchase, error) {
	// Ambil purchase
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, fmt.Errorf("purchase tidak ditemukan")
	}

	// Validasi kepemilikan
	if purchase.UserID == nil || *purchase.UserID != userID {
		return nil, fmt.Errorf("akses ditolak: bukan milik Anda")
	}

	// Validasi status harus pending
	if purchase.PaymentStatus != models.Pending {
		return nil, fmt.Errorf("pembayaran tidak bisa diproses, status saat ini: %s", purchase.PaymentStatus)
	}

	// Cek apakah sudah expired berdasarkan ExpiredAt
	if purchase.ExpiredAt != nil && time.Now().After(*purchase.ExpiredAt) {
		return nil, fmt.Errorf("pembayaran tidak bisa diproses karena transaksi sudah kedaluwarsa")
	}

	// Update status & bukti bayar
	purchase.PaymentProof = &proofURL
	purchase.PaymentStatus = models.WaitingConfirmation
	purchase.UpdatedAt = time.Now()

	if err := s.purchaseRepo.Update(purchase); err != nil {
		return nil, err
	}

	purchaseWithPrice, err := s.purchaseRepo.GetPurchaseByID(purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil ulang purchase: %w", err)
	}

	return purchaseWithPrice, nil

}

// CancelPurchase is using for cancel purchase
func (s *PurchaseService) CancelPurchase(userID, purchaseID uuid.UUID) (*models.Purchase, error) {
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, fmt.Errorf("purchase tidak ditemukan")
	}

	// Cek kepemilikan
	if purchase.UserID == nil || *purchase.UserID != userID {
		return nil, fmt.Errorf("akses ditolak: bukan milik Anda")
	}

	// Cek status valid untuk cancel
	if purchase.PaymentStatus != models.Pending && purchase.PaymentStatus != models.WaitingConfirmation {
		return nil, fmt.Errorf("tidak bisa membatalkan purchase dengan status: %s", purchase.PaymentStatus)
	}

	// Set status cancelled
	purchase.PaymentStatus = models.Cancelled
	purchase.UpdatedAt = time.Now()

	err = s.purchaseRepo.Update(purchase)
	if err != nil {
		return nil, err
	}

	purchaseWithPrice, err := s.purchaseRepo.GetPurchaseByID(purchase.ID)
	if err != nil {
		return nil, fmt.Errorf("Gagal mengambil ulang purchase: %w", err)
	}

	return purchaseWithPrice, nil

}
