package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CertificateController handles certificate endpoints
type CertificateController struct {
	certificateService services.ICertificateService
}

// NewCertificateController creates a new instance
func NewCertificateController(certService services.ICertificateService) *CertificateController {
	return &CertificateController{
		certificateService: certService,
	}
}

// GenerateCertificate generates a certificate for the authenticated student in a batch
func (ctrl *CertificateController) GenerateCertificate(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchIDParam := c.Params("batchID")
	if batchIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "batchID is required", "")
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid batchID", err.Error())
	}

	// Ambil user dari context
	userClaims := c.Locals("user").(*utils.Claims)

	// Panggil service untuk generate certificate
	cert, err := ctrl.certificateService.EnsureCertificate(ctx, batchID, userClaims)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Failed to generate certificate", err.Error())
	}

	// Map ke DTO
	var certResponse dto.CertificateResponse
	if copyErr := copier.Copy(&certResponse, cert); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map certificate data", copyErr.Error())
	}

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Certificate generated successfully", certResponse, nil)
}

// GetCertificate retrieves the certificate for the authenticated student in a batch
func (ctrl *CertificateController) GetCertificate(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchIDParam := c.Params("batchID")
	if batchIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "batchID is required", "")
	}

	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid batchID", err.Error())
	}

	// Ambil user dari context
	userClaims := c.Locals("user").(*utils.Claims)

	// Panggil service untuk get certificate
	cert, err := ctrl.certificateService.GetCertificateByBatchUser(ctx, batchID, userClaims)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Certificate not found", "")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve certificate", err.Error())
	}

	// Map ke DTO
	var certResponse dto.CertificateResponse
	if copyErr := copier.Copy(&certResponse, cert); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map certificate data", copyErr.Error())
	}

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Certificate retrieved successfully", certResponse, nil)
}

// GetBatchCertificates retrieves all certificates in a batch
func (ctrl *CertificateController) GetBatchCertificates(c *fiber.Ctx) error {
	ctx := c.UserContext()

	batchIDParam := c.Params("batchID")
	if batchIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "batchID is required", "")
	}
	batchID, err := uuid.Parse(batchIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid batchID", err.Error())
	}

	// Ambil user dari context
	userClaims := c.Locals("user").(*utils.Claims)

	// Call service
	certs, err := ctrl.certificateService.GetCertificatesByBatch(ctx, batchID, userClaims)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve certificates", err.Error())
	}

	// Map ke DTO slice
	var certResponses []dto.CertificateResponse
	if copyErr := copier.Copy(&certResponses, certs); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map certificate data", copyErr.Error())
	}

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Certificates retrieved successfully", certResponses, nil)
}

// GetBatchCertificate retrieves detail of a certificate in a batch
func (ctrl *CertificateController) GetBatchCertificate(c *fiber.Ctx) error {
	ctx := c.UserContext()

	// Parse certificateID
	certIDParam := c.Params("certificateID")
	if certIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "certificateID is required", "")
	}
	certID, err := uuid.Parse(certIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid certificateID", err.Error())
	}

	// Ambil user dari context
	userClaims := c.Locals("user").(*utils.Claims)

	// Call service
	cert, err := ctrl.certificateService.GetCertificateDetail(ctx, certID, userClaims)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Certificate not found", "")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve certificate detail", err.Error())
	}

	// Map ke DTO
	var certResponse dto.CertificateResponse
	if copyErr := copier.Copy(&certResponse, cert); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map certificate data", copyErr.Error())
	}

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Certificate detail retrieved successfully", certResponse, nil)
}

// VerifyCertificate verifies the authenticity of a certificate (public)
func (ctrl *CertificateController) VerifyCertificate(c *fiber.Ctx) error {
	ctx := c.UserContext()

	certIDParam := c.Params("certificateID")
	if certIDParam == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "certificateID is required", "")
	}

	certID, err := uuid.Parse(certIDParam)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid certificateID", err.Error())
	}

	// Service: langsung ambil by ID (tanpa claims)
	cert, err := ctrl.certificateService.VerifyCertificate(ctx, certID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Certificate not found", "")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to verify certificate", err.Error())
	}

	// Map ke DTO
	var certResponse dto.CertificateResponse
	if copyErr := copier.Copy(&certResponse, cert); copyErr != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to map certificate data", copyErr.Error())
	}

	return utils.SuccessWithMeta(c, fiber.StatusOK, "Certificate is valid", certResponse, nil)
}
