package controllers

import (
	"brevet-api/config"
	"brevet-api/dto"
	"brevet-api/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UploadDocument handles document upload
func UploadDocument(c *fiber.Ctx) error {
	// Ambil DTO dari c.Locals
	body := c.Locals("body").(*dto.UploadRequest)

	location := filepath.Clean(body.Location)

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	if !utils.IsAllowedExtension(file.Filename, utils.AllowedDocumentExtensions) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Hanya file dokumen yang diperbolehkan", nil)
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext

	baseDir := config.GetEnv("UPLOAD_DIR", "./public/uploads")

	saveDir := filepath.Join(baseDir, location)

	if !utils.IsSafePath(baseDir, saveDir) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Lokasi upload tidak valid", nil)
	}
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat folder upload", err.Error())
	}

	savePath := filepath.Join(saveDir, filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan file", err.Error())
	}

	url := fmt.Sprintf("/uploads/%s/%s", location, filename)

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// UploadImage handles image upload
func UploadImage(c *fiber.Ctx) error {
	body := c.Locals("body").(*dto.UploadRequest)

	location := filepath.Clean(body.Location)

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File tidak ditemukan", err.Error())
	}

	if !utils.IsAllowedExtension(file.Filename, utils.AllowedImageExtensions) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Hanya file gambar yang diperbolehkan", nil)
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext

	baseDir := config.GetEnv("UPLOAD_DIR", "./public/uploads")

	saveDir := filepath.Join(baseDir, location)

	if !utils.IsSafePath(baseDir, saveDir) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Lokasi upload tidak valid", nil)
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat folder upload", err.Error())
	}

	savePath := filepath.Join(saveDir, filename)
	if err := c.SaveFile(file, savePath); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menyimpan file", err.Error())
	}

	url := fmt.Sprintf("/uploads/%s/%s", location, filename)

	return utils.SuccessResponse(c, fiber.StatusOK, "Upload berhasil", url)
}

// DeleteFile handles file deletion
func DeleteFile(c *fiber.Ctx) error {
	// Ambil body dari c.Locals, pastikan sudah validasi dan parsing di middleware
	body := c.Locals("body").(*dto.DeleteRequest)

	cleanPath := filepath.Clean(body.FilePath)

	// Validasi path menggunakan helper
	targetPath, err := utils.IsSafeDeletePath(cleanPath)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memverifikasi path", err.Error())
	}
	if targetPath == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "File path tidak valid", nil)
	}

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "File tidak ditemukan", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus file", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "File berhasil dihapus", nil)
}
