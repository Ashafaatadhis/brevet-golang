package services

import (
	"brevet-api/dto"
	"brevet-api/helpers"
	"brevet-api/models"
	"brevet-api/repository"
	"os"
	"path/filepath"
	"runtime"

	"brevet-api/utils"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nguyenthenguyen/docx"
	"gorm.io/gorm"
)

// PurchaseService provides methods for managing purchases
type PurchaseService struct {
	purchaseRepo *repository.PurchaseRepository
	userRepo     *repository.UserRepository
	batchRepo    *repository.BatchRepository
	emailService *EmailService
	db           *gorm.DB
}

// NewPurchaseService creates a new instance of PurchaseService
func NewPurchaseService(purchaseRepository *repository.PurchaseRepository, batchRepo *repository.BatchRepository, emailService *EmailService, db *gorm.DB) *PurchaseService {
	return &PurchaseService{purchaseRepo: purchaseRepository, batchRepo: batchRepo, emailService: emailService, db: db}
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

func (s *PurchaseService) generateAndSendReceipt(purchase *models.Purchase) error {
	// 1. Siapkan data placeholder
	total := int(purchase.TransferAmount) + purchase.UniqueCode
	var nama, npm, kelas, atasNama string

	if purchase.User != nil {
		nama = purchase.User.Name
		if purchase.User.Profile != nil {
			if purchase.User.Profile.NIM.Valid {
				npm = purchase.User.Profile.NIM.String
			} else {
				npm = "-"
			}
			if purchase.User.Profile.GroupType != nil {
				kelas = helpers.FormatGroupType(string(*purchase.User.Profile.GroupType))
			} else {
				kelas = "-"
			}
		} else {
			npm = "-"
			kelas = "-"
		}
	} else {
		nama = "-"
		npm = "-"
		kelas = "-"
	}

	if purchase.BuyerBankAccountName != nil {
		atasNama = *purchase.BuyerBankAccountName
	} else {
		atasNama = "-"
	}

	data := map[string]string{
		"{{NOMOR}}":     fmt.Sprintf("%07d", purchase.InvoiceNumber),
		"{{NAMA}}":      nama,
		"{{NPM}}":       npm,
		"{{KELAS}}":     kelas,
		"{{JUMLAH}}":    fmt.Sprintf("%s", helpers.FormatWithDot(total)),
		"{{ATASNAMA}}":  atasNama,
		"{{TERBILANG}}": strings.TrimSpace(helpers.NumToString(int(total))) + " Rupiah",
	}

	// Buat folder temp unik untuk simpan file ini
	tempDir := os.TempDir()
	uniqueFolder := filepath.Join(tempDir, "kwitansi-"+uuid.New().String())
	if err := os.MkdirAll(uniqueFolder, 0755); err != nil {
		return fmt.Errorf("gagal buat folder temp: %w", err)
	}

	fmt.Println(tempDir, "faefeaw")
	// Pastikan folder dan isinya dihapus setelah selesai
	defer os.RemoveAll(uniqueFolder)

	// Baca template dan replace placeholder
	r, err := docx.ReadDocxFile("templates/contoh_blanko.docx")
	if err != nil {
		return fmt.Errorf("baca template gagal: %w", err)
	}
	doc := r.Editable()
	for k, v := range data {
		doc.Replace(k, v, -1)
	}

	outputDocx := filepath.Join(uniqueFolder, fmt.Sprintf("kwitansi_%07d.docx", purchase.InvoiceNumber))
	if err := doc.WriteToFile(outputDocx); err != nil {
		return fmt.Errorf("simpan docx gagal: %w", err)
	}

	// Convert DOCX ke PDF
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("soffice", "--headless", "--convert-to", "pdf", outputDocx, "--outdir", uniqueFolder)
	} else {
		cmd = exec.Command(
			"libreoffice",
			"--headless",
			"--convert-to", "pdf:writer_pdf_Export",
			outputDocx,
			"--outdir", uniqueFolder,
		)

	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("convert pdf gagal: %w", err)
	}

	outputPDF := strings.Replace(outputDocx, ".docx", ".pdf", 1)

	// Kirim email dengan attachment PDF
	var email string
	if purchase.User != nil {
		email = purchase.User.Email
	}
	if email == "" {
		return fmt.Errorf("email user tidak tersedia")
	}
	if err := s.emailService.SendWithAttachment(
		email,
		"Kwitansi Pembayaran",
		"Terima kasih, pembayaran Anda telah diterima. Terlampir kwitansi.",
		outputPDF,
	); err != nil {
		return fmt.Errorf("kirim email gagal: %w", err)
	}

	return nil
}

// GetPaidBatchIDs for get all batch where user has paid
func (s *PurchaseService) GetPaidBatchIDs(userID string) ([]string, error) {
	return s.purchaseRepo.GetPaidBatchIDs(userID)
}

// CreatePurchase is for create purchase
func (s *PurchaseService) CreatePurchase(userID uuid.UUID, batchID uuid.UUID) (*models.Purchase, error) {
	var result *models.Purchase

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		purchaseRepo := s.purchaseRepo.WithTx(tx)
		userRepo := s.userRepo.WithTx(tx)

		// 1. Cek apakah sudah pernah beli (pakai tx)
		hasPaid, err := purchaseRepo.HasPurchaseWithStatus(userID, batchID,
			[]models.PaymentStatus{
				models.Pending, models.WaitingConfirmation, models.Paid,
			}...,
		)
		if err != nil {
			return err
		}
		if hasPaid {
			return errors.New("Anda sudah memiliki transaksi untuk batch ini")
		}

		// 2. Ambil user
		user, err := userRepo.FindByID(userID)
		if err != nil {
			return fmt.Errorf("User tidak ditemukan: %w", err)
		}
		if user.Profile == nil || user.Profile.GroupType == nil {
			return fmt.Errorf("User belum memiliki GroupType yang valid")
		}

		// this leak exist when user group type is not verified by admin
		allowed, err := purchaseRepo.IsGroupTypeAllowedForBatch(batchID, *user.Profile.GroupType)
		if err != nil {
			return fmt.Errorf("gagal validasi group type batch: %w", err)
		}
		if !allowed {
			return fmt.Errorf("Batch ini tidak tersedia untuk GroupType '%s'", *user.Profile.GroupType)
		}

		// 3. Ambil harga
		price, err := purchaseRepo.GetPriceByGroupType(user.Profile.GroupType)
		if err != nil {
			return fmt.Errorf("harga untuk group_type '%s' tidak ditemukan: %w", *user.Profile.GroupType, err)
		}

		// 4. Buat purchase
		expiredAt := time.Now().Add(24 * time.Hour)
		uniqueCode := utils.GenerateUniqueCode()
		transferAmount := price.Price + float64(uniqueCode)
		purchase := &models.Purchase{
			UserID:         &userID,
			BatchID:        &batchID,
			UniqueCode:     uniqueCode,
			TransferAmount: transferAmount,
			PriceID:        price.ID,
			ExpiredAt:      &expiredAt,
			PaymentStatus:  models.Pending,
		}
		if err := purchaseRepo.Create(purchase); err != nil {
			return err
		}

		// 5. Ambil ulang setelah insert (pakai tx juga)
		result, err = purchaseRepo.GetPurchaseByID(purchase.ID)
		if err != nil {
			return fmt.Errorf("Gagal mengambil ulang purchase: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateStatusPayment verification payment service
func (s *PurchaseService) UpdateStatusPayment(purchaseID uuid.UUID, body *dto.UpdateStatusPayment) (*models.Purchase, error) {
	var result *models.Purchase

	err := utils.WithTransaction(s.db, func(tx *gorm.DB) error {
		purchaseRepo := s.purchaseRepo.WithTx(tx)
		batchRepo := s.batchRepo.WithTx(tx).WithLock()

		purchase, err := purchaseRepo.GetPurchaseByID(purchaseID)
		if err != nil {
			return fmt.Errorf("data tidak ditemukan: %w", err)
		}

		// if purchase.PaymentStatus != models.WaitingConfirmation {
		// 	return fmt.Errorf("status pembayaran tidak bisa diverifikasi")
		// }

		if body.PaymentStatus == models.Paid {
			batch, err := batchRepo.FindByID(*purchase.BatchID)
			if err != nil {
				return fmt.Errorf("batch tidak ditemukan: %w", err)
			}

			count, err := purchaseRepo.CountPaidByBatchID(*purchase.BatchID)
			if err != nil {
				return fmt.Errorf("gagal menghitung paid: %w", err)
			}

			if int(count) >= batch.Quota {
				return fmt.Errorf("kuota batch sudah penuh")
			}

			go func(purchase *models.Purchase) {
				if err := s.generateAndSendReceipt(purchase); err != nil {
					log.Printf("gagal mengirim kwitansi: %v", err)
				}
			}(purchase)
		}

		purchase.PaymentStatus = body.PaymentStatus
		if err := purchaseRepo.Update(purchase); err != nil {
			return fmt.Errorf("gagal update status: %w", err)
		}

		result, err = s.purchaseRepo.WithTx(tx).GetPurchaseByID(purchase.ID)
		if err != nil {
			return fmt.Errorf("gagal mengambil ulang purchase: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// PayPurchase is for pay purchase
func (s *PurchaseService) PayPurchase(userID uuid.UUID, purchaseID uuid.UUID, body *dto.PayPurchaseRequest) (*models.Purchase, error) {
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
	purchase.PaymentProof = &body.PaymentProofURL
	purchase.PaymentStatus = models.WaitingConfirmation
	purchase.BuyerBankAccountName = &body.BuyerBankAccountName
	purchase.BuyerBankAccountNumber = &body.BuyerBankAccountNumber
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
