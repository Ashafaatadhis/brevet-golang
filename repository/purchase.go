package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPurchaseRepository interface
type IPurchaseRepository interface {
	WithTx(tx *gorm.DB) IPurchaseRepository
	GetAllFilteredPurchases(ctx context.Context, opts utils.QueryOptions) ([]models.Purchase, int64, error)
	GetMyFilteredPurchases(ctx context.Context, opts utils.QueryOptions, userID uuid.UUID) ([]models.Purchase, int64, error)
	GetPurchaseByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error)
	HasPaid(ctx context.Context, userID uuid.UUID, batchID uuid.UUID) (bool, error)
	CountPaidByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error)
	HasPurchaseWithStatus(ctx context.Context, userID uuid.UUID, batchID uuid.UUID, statuses ...models.PaymentStatus) (bool, error)
	GetPaidBatchIDs(ctx context.Context, userID string) ([]string, error)
	Create(ctx context.Context, purchase *models.Purchase) error
	GetPriceByGroupType(ctx context.Context, groupType *models.GroupType) (*models.Price, error)
	Update(ctx context.Context, course *models.Purchase) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error)
	IsGroupTypeAllowedForBatch(ctx context.Context, batchID uuid.UUID, groupType models.GroupType) (bool, error)
}

// PurchaseRepository is a struct that represents a purchase repository
type PurchaseRepository struct {
	db *gorm.DB
}

// NewPurchaseRepository creates a new purchase repository
func NewPurchaseRepository(db *gorm.DB) IPurchaseRepository {
	return &PurchaseRepository{db: db}
}

// WithTx running with transaction
func (r *PurchaseRepository) WithTx(tx *gorm.DB) IPurchaseRepository {
	return &PurchaseRepository{db: tx}
}

// GetAllFilteredPurchases retrieves all purchases with pagination and filtering options
func (r *PurchaseRepository) GetAllFilteredPurchases(ctx context.Context, opts utils.QueryOptions) ([]models.Purchase, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Purchase{}, &models.User{}, &models.Batch{}, &models.Price{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Purchase{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	db = utils.ApplyFiltersWithJoins(db, "purchases", opts.Filters, validSortFields, joinConditions, joinedRelations)

	var total int64
	db.Count(&total)

	var purchases []models.Purchase
	err := db.Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("User").
		Preload("Batch").
		Preload("Price").
		Find(&purchases).Error

	return purchases, total, err
}

// GetMyFilteredPurchases retrieves all purchases with pagination and filtering options
func (r *PurchaseRepository) GetMyFilteredPurchases(ctx context.Context, opts utils.QueryOptions, userID uuid.UUID) ([]models.Purchase, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Purchase{}, &models.User{}, &models.Batch{}, &models.Price{})

	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.WithContext(ctx).Model(&models.Purchase{})

	joinConditions := map[string]string{}
	joinedRelations := map[string]bool{}

	// filter dinamis (misal berdasarkan batch name, price, dll)
	db = utils.ApplyFiltersWithJoins(db, "purchases", opts.Filters, validSortFields, joinConditions, joinedRelations)

	// ✅ filter khusus untuk user_id
	db = db.Where("purchases.user_id = ?", userID)

	// hitung total setelah semua filter diterapkan
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var purchases []models.Purchase
	err := db.
		Order(fmt.Sprintf("%s %s", sort, order)).
		Limit(opts.Limit).
		Offset(opts.Offset).
		Preload("User").
		Preload("Batch").
		Preload("Price").
		Find(&purchases).Error

	return purchases, total, err
}

// GetPurchaseByID is for get purchase by id
func (r *PurchaseRepository) GetPurchaseByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.WithContext(ctx).Preload("User").Preload("User.Profile").
		Preload("Batch").
		Preload("Price").First(&purchase, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

// HasPaid check if user has paid in this batch by batchid
func (r *PurchaseRepository) HasPaid(ctx context.Context, userID uuid.UUID, batchID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Purchase{}).
		Where("user_id = ? AND batch_id = ? AND payment_status = ?", userID, batchID, models.Paid).
		Count(&count).Error
	return count > 0, err
}

// CountPaidByBatchID retrieves the count of paid purchases for a specific batch
func (r *PurchaseRepository) CountPaidByBatchID(ctx context.Context, batchID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Purchase{}).
		Where("batch_id = ? AND payment_status = ?", batchID, models.Paid).
		Count(&count).Error
	return count, err
}

// HasPurchaseWithStatus check if user has in status
func (r *PurchaseRepository) HasPurchaseWithStatus(ctx context.Context, userID uuid.UUID, batchID uuid.UUID, statuses ...models.PaymentStatus) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Purchase{}).
		Where("user_id = ? AND batch_id = ? AND payment_status IN ?", userID, batchID, statuses).
		Count(&count).Error
	return count > 0, err
}

// GetPaidBatchIDs get all batch where the user has paid
func (r *PurchaseRepository) GetPaidBatchIDs(ctx context.Context, userID string) ([]string, error) {
	var batchIDs []string
	err := r.db.WithContext(ctx).Model(&models.Purchase{}).
		Where("user_id = ? AND status = ?", userID, "paid").
		Pluck("batch_id", &batchIDs).Error
	return batchIDs, err
}

// Create creates a new purchase
func (r *PurchaseRepository) Create(ctx context.Context, purchase *models.Purchase) error {
	return r.db.WithContext(ctx).Create(purchase).Error
}

// GetPriceByGroupType get price by group type
func (r *PurchaseRepository) GetPriceByGroupType(ctx context.Context, groupType *models.GroupType) (*models.Price, error) {
	var price models.Price
	if err := r.db.WithContext(ctx).Where("group_type = ?", groupType).First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

// Update updates an existing purchase
func (r *PurchaseRepository) Update(ctx context.Context, course *models.Purchase) error {
	return r.db.WithContext(ctx).Save(course).Error
}

// FindByID is repo for find purchase by id
func (r *PurchaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.WithContext(ctx).First(&purchase, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

// IsGroupTypeAllowedForBatch for checking group type
func (r *PurchaseRepository) IsGroupTypeAllowedForBatch(ctx context.Context, batchID uuid.UUID, groupType models.GroupType) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.BatchGroup{}).
		Where("batch_id = ? AND group_type = ?", batchID, groupType).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
