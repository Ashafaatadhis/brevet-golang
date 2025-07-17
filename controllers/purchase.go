package controllers

import (
	"brevet-api/dto"
	"brevet-api/models"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// PurchaseController handles purchase-related operations
type PurchaseController struct {
	purchaseService *services.PurchaseService
	db              *gorm.DB
}

// NewPurchaseController creates a new NewPurchaseController
func NewPurchaseController(purchaseService *services.PurchaseService, db *gorm.DB) *PurchaseController {
	return &PurchaseController{
		purchaseService: purchaseService,
		db:              db,
	}
}

// GetAllPurchases retrieves a list of purchases with pagination and filtering options
func (ctrl *PurchaseController) GetAllPurchases(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)

	purchases, total, err := ctrl.purchaseService.GetAllFilteredPurchases(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch purchases", err.Error())
	}

	var purchasesResponse []dto.PurchaseResponse
	if copyErr := copier.Copy(&purchasesResponse, purchases); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Purchases fetched", purchasesResponse, meta)
}

// GetMyPurchase retrieves a list of purchases with pagination and filtering options
func (ctrl *PurchaseController) GetMyPurchase(c *fiber.Ctx) error {
	opts := utils.ParseQueryOptions(c)
	user := c.Locals("user").(*utils.Claims)

	// Kalau role siswa, paksa filter user_id = dirinya sendiri
	if user.Role == string(models.RoleTypeSiswa) {
		opts.Filters["user_id"] = user.UserID.String()
	}

	purchases, total, err := ctrl.purchaseService.GetAllFilteredPurchases(opts)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch purchases", err.Error())
	}

	var purchasesResponse []dto.PurchaseResponse
	if copyErr := copier.Copy(&purchasesResponse, purchases); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	meta := utils.BuildPaginationMeta(total, opts.Limit, opts.Page)
	return utils.SuccessWithMeta(c, fiber.StatusOK, "Purchases fetched", purchasesResponse, meta)
}

// GetMyPurchaseByID retrieves a course by its slug (ID)
func (ctrl *PurchaseController) GetMyPurchaseByID(c *fiber.Ctx) error {

	user := c.Locals("user").(*utils.Claims)

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	purchase, err := ctrl.purchaseService.GetPurchaseByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Purchase Doesn't Exist", err.Error())
	}

	// // Jika siswa, pastikan milik sendiri
	if user.Role == string(models.RoleTypeSiswa) && purchase.UserID != nil && *purchase.UserID != user.UserID {
		return utils.ErrorResponse(c, fiber.StatusForbidden, "Akses ditolak", "Purchase ini bukan milik Anda")
	}

	var purchaseResponse dto.PurchaseResponse
	if copyErr := copier.Copy(&purchaseResponse, purchase); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Purchase fetched", purchaseResponse)
}

// GetPurchaseByID retrieves a course by its slug (ID)
func (ctrl *PurchaseController) GetPurchaseByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)

	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid UUID format", err.Error())
	}

	purchase, err := ctrl.purchaseService.GetPurchaseByID(id)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Purchase Doesn't Exist", err.Error())
	}

	var purchaseResponse dto.PurchaseResponse
	if copyErr := copier.Copy(&purchaseResponse, purchase); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Purchase fetched", purchaseResponse)
}

// CreatePurchase creates a new purchase
func (ctrl *PurchaseController) CreatePurchase(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.CreatePurchase)
	user := c.Locals("user").(*utils.Claims)

	purchase, err := ctrl.purchaseService.CreatePurchase(user.UserID, body.BatchID)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to create purchase", err.Error())
	}

	var purchaseResponse dto.PurchaseResponse
	if copyErr := copier.Copy(&purchaseResponse, purchase); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 201, "Purchase created successfully", purchaseResponse)
}

// UpdateStatusPayment untuk verify pembayaran
func (ctrl *PurchaseController) UpdateStatusPayment(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", err.Error())
	}

	body := c.Locals("body").(*dto.UpdateStatusPayment)

	purchase, err := ctrl.purchaseService.UpdateStatusPayment(id, body)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Gagal verifikasi pembayaran", err.Error())
	}

	var purchaseResponse dto.PurchaseResponse
	if copyErr := copier.Copy(&purchaseResponse, purchase); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Pembayaran berhasil diverifikasi", purchaseResponse)
}

// Pay is controller for paying purchase
func (ctrl *PurchaseController) Pay(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.PayPurchaseRequest)
	user := c.Locals("user").(*utils.Claims)
	purchaseIDStr := c.Params("id")

	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid purchase ID", err.Error())
	}

	// Panggil service untuk proses bayar
	purchase, err := ctrl.purchaseService.PayPurchase(user.UserID, purchaseID, body.PaymentProofURL)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Failed to upload payment proof", err.Error())
	}

	var purchaseResponse dto.PurchaseResponse
	if copyErr := copier.Copy(&purchaseResponse, purchase); copyErr != nil {
		return utils.ErrorResponse(c, 500, "Failed to map purchase data", copyErr.Error())
	}

	return utils.SuccessResponse(c, 200, "Payment proof uploaded successfully", purchaseResponse)
}

// Cancel is controller for cancel purchase
func (ctrl *PurchaseController) Cancel(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.Claims)
	purchaseIDStr := c.Params("id")

	purchaseID, err := uuid.Parse(purchaseIDStr)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid purchase ID", err.Error())
	}

	purchase, err := ctrl.purchaseService.CancelPurchase(user.UserID, purchaseID)
	if err != nil {
		return utils.ErrorResponse(c, 400, "Gagal membatalkan pembelian", err.Error())
	}

	var response dto.PurchaseResponse
	if err := copier.Copy(&response, purchase); err != nil {
		return utils.ErrorResponse(c, 500, "Gagal memetakan data", err.Error())
	}

	return utils.SuccessResponse(c, 200, "Pembelian berhasil dibatalkan", response)
}
