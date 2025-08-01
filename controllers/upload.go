package controllers

import (
	"brevet-api/dto"
	"brevet-api/helpers"
	"brevet-api/services"
	"brevet-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
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
	log := helpers.LoggerFromCtx(c.UserContext())
	log.Info("UploadImage handler called")
	body := c.Locals("body").(*dto.UploadRequest)
	file, err := c.FormFile("file")
	if err != nil {
		log.WithError(err).Warn("File gambar tidak ditemukan di request")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	url, err := fc.fileService.SaveFile(c, file, body.Location, utils.AllowedImageExtensions)
	if err != nil {
		log.WithError(err).Warn("Gagal menyimpan file gambar")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	log.WithFields(logrus.Fields{
		"file": file.Filename,
		"url":  url,
	}).Info("Upload gambar berhasil")

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// UploadDocument handles document file uploads
func (fc *FileController) UploadDocument(c *fiber.Ctx) error {

	log := helpers.LoggerFromCtx(c.UserContext())

	log.Info("UploadDocument handler called")

	body := c.Locals("body").(*dto.UploadRequest)
	file, err := c.FormFile("file")
	if err != nil {
		log.WithError(err).Warn("File dokumen tidak ditemukan di request")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	url, err := fc.fileService.SaveFile(c, file, body.Location, utils.AllowedDocumentExtensions)
	if err != nil {
		log.WithError(err).Warn("Gagal menyimpan file dokumen")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	log.WithFields(logrus.Fields{
		"file": file.Filename,
		"url":  url,
	}).Info("Upload dokumen berhasil")

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// DeleteFile handles file deletion
func (fc *FileController) DeleteFile(c *fiber.Ctx) error {
	log := helpers.LoggerFromCtx(c.UserContext())

	log.Info("DeleteFile handler called")

	body := c.Locals("body").(*dto.DeleteRequest)
	if err := fc.fileService.DeleteFile(body.FilePath); err != nil {
		log.WithError(err).Warn("Gagal menghapus file")
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	log.WithField("file_path", body.FilePath).Info("File berhasil dihapus")
	return utils.SuccessResponse(c, fiber.StatusOK, "File berhasil dihapus", nil)
}
