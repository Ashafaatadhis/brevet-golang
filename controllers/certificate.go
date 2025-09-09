package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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
	cert, err := ctrl.certificateService.EnsureCertificate(ctx, batchID, userClaims.UserID)
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
