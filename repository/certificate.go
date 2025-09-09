package repository

import (
	"brevet-api/models"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ICertificateRepository interface
type ICertificateRepository interface {
	GetByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (*models.Certificate, error)
	Create(ctx context.Context, cert *models.Certificate) error
	Update(ctx context.Context, cert *models.Certificate) error
	GetLastSequenceByBatch(ctx context.Context, batchID uuid.UUID) (int, error)
}

// CertificateRepository is a struct that represents a certificate repository
type CertificateRepository struct {
	db *gorm.DB
}

// NewCertificateRepository creates a new certificate repository
func NewCertificateRepository(db *gorm.DB) ICertificateRepository {
	return &CertificateRepository{db: db}
}

// GetByBatchUser retrieves a certificate by batch ID and user ID
func (r *CertificateRepository) GetByBatchUser(ctx context.Context, batchID, userID uuid.UUID) (*models.Certificate, error) {
	var cert models.Certificate
	err := r.db.WithContext(ctx).
		Where("batch_id = ? AND user_id = ?", batchID, userID).
		First(&cert).Error
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// Create inserts a new certificate record
func (r *CertificateRepository) Create(ctx context.Context, cert *models.Certificate) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

// Update updates an existing certificate in the database
func (r *CertificateRepository) Update(ctx context.Context, cert *models.Certificate) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

// GetLastSequenceByBatch get last
func (r *CertificateRepository) GetLastSequenceByBatch(ctx context.Context, batchID uuid.UUID) (int, error) {
	var lastNumber string
	err := r.db.WithContext(ctx).
		Model(&models.Certificate{}).
		Where("batch_id = ?", batchID).
		Order("created_at DESC").
		Pluck("number", &lastNumber).Error
	if err != nil {
		return 0, err
	}
	if lastNumber == "" {
		return 0, nil
	}

	// parsing nomor terakhir → ambil bagian urutannya
	var seq int
	_, err = fmt.Sscanf(lastNumber, "20100112-%*d %d", &seq)
	if err != nil {
		return 0, nil
	}
	return seq, nil
}
