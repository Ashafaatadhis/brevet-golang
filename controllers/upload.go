package controllers

import (
	"brevet-api/dto"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
)

// FileController handles file upload and deletion operations
type FileController struct {
	fileService *services.FileService
}

// NewFileController creates a new FileController instance
func NewFileController(fileService *services.FileService) *FileController {
	return &FileController{fileService: fileService}
}

// UploadImage handles image file uploads
func (fc *FileController) UploadImage(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.UploadRequest)
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	url, err := fc.fileService.SaveFile(c, file, body.Location, utils.AllowedImageExtensions)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// UploadDocument handles document file uploads
func (fc *FileController) UploadDocument(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.UploadRequest)
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	url, err := fc.fileService.SaveFile(c, file, body.Location, utils.AllowedDocumentExtensions)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// DeleteFile handles file deletion
func (fc *FileController) DeleteFile(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.DeleteRequest)
	if err := fc.fileService.DeleteFile(body.FilePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "File berhasil dihapus", nil)
}
