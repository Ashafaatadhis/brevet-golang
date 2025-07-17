package repository

import (
	"brevet-api/models"
	"brevet-api/utils"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PurchaseRepository is a struct that represents a purchase repository
type PurchaseRepository struct {
	db *gorm.DB
}

// NewPurchaseRepository creates a new purchase repository
func NewPurchaseRepository(db *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

// WithTx running with transaction
func (r *PurchaseRepository) WithTx(tx *gorm.DB) *PurchaseRepository {
	return &PurchaseRepository{db: tx}
}

// GetAllFilteredPurchases retrieves all purchases with pagination and filtering options
func (r *PurchaseRepository) GetAllFilteredPurchases(opts utils.QueryOptions) ([]models.Purchase, int64, error) {
	validSortFields := utils.GetValidColumnsFromStruct(&models.Purchase{}, &models.User{}, &models.Batch{}, &models.Price{})
	fmt.Print(validSortFields, "TAI")
	sort := opts.Sort
	if !validSortFields[sort] {
		sort = "id"
	}

	order := opts.Order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	db := r.db.Model(&models.Purchase{})

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

// GetPurchaseByID is for get purchase by id
func (r *PurchaseRepository) GetPurchaseByID(id uuid.UUID) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.Preload("User").
		Preload("Batch").
		Preload("Price").First(&purchase, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}

// HasPaid check if user has paid in this batch by batchid
func (r *PurchaseRepository) HasPaid(userID uuid.UUID, batchID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).
		Where("user_id = ? AND batch_id = ? AND payment_status = ?", userID, batchID, models.Paid).
		Count(&count).Error
	return count > 0, err
}

// HasPurchaseWithStatus check if user has in status
func (r *PurchaseRepository) HasPurchaseWithStatus(userID uuid.UUID, batchID uuid.UUID, statuses ...models.PaymentStatus) (bool, error) {
	var count int64
	err := r.db.Model(&models.Purchase{}).
		Where("user_id = ? AND batch_id = ? AND payment_status IN ?", userID, batchID, statuses).
		Count(&count).Error
	return count > 0, err
}

// GetPaidBatchIDs get all batch where the user has paid
func (r *PurchaseRepository) GetPaidBatchIDs(userID string) ([]string, error) {
	var batchIDs []string
	err := r.db.Model(&models.Purchase{}).
		Where("user_id = ? AND status = ?", userID, "paid").
		Pluck("batch_id", &batchIDs).Error
	return batchIDs, err
}

// Create creates a new purchase
func (r *PurchaseRepository) Create(purchase *models.Purchase) error {
	return r.db.Create(purchase).Error
}

// GetPriceByGroupType get price by group type
func (r *PurchaseRepository) GetPriceByGroupType(groupType *models.GroupType) (*models.Price, error) {
	var price models.Price
	if err := r.db.Where("group_type = ?", groupType).First(&price).Error; err != nil {
		return nil, err
	}
	return &price, nil
}

// Update updates an existing purchase
func (r *PurchaseRepository) Update(course *models.Purchase) error {
	return r.db.Save(course).Error
}

// FindByID is repo for find purchase by id
func (r *PurchaseRepository) FindByID(id uuid.UUID) (*models.Purchase, error) {
	var purchase models.Purchase
	err := r.db.First(&purchase, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &purchase, nil
}
