package services

import (
	"brevet-api/config"
	"brevet-api/utils"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// FileService is a struct that represents a file service
type FileService struct {
	BaseDir string
}

// NewFileService creates a new file service
func NewFileService() *FileService {
	return &FileService{
		BaseDir: config.GetEnv("UPLOAD_DIR", "./public/uploads"),
	}
}

// SaveFile saves an uploaded file to the specified location with validation for allowed extensions
func (s *FileService) SaveFile(ctx *fiber.Ctx, file *multipart.FileHeader, location string, allowedExts []string) (string, error) {
	if !utils.IsAllowedExtension(file.Filename, allowedExts) {
		return "", fmt.Errorf("Ekstensi file tidak diperbolehkan")
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	saveDir := filepath.Join(s.BaseDir, filepath.Clean(location))

	if !utils.IsSafePath(s.BaseDir, saveDir) {
		return "", fmt.Errorf("Lokasi upload tidak valid")
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("Gagal membuat folder upload: %w", err)
	}

	savePath := filepath.Join(saveDir, filename)
	if err := ctx.SaveFile(file, savePath); err != nil {
		return "", fmt.Errorf("Gagal menyimpan file: %w", err)
	}

	return fmt.Sprintf("/uploads/%s/%s", location, filename), nil
}

// DeleteFile deletes a file from the server after validating the path
func (s *FileService) DeleteFile(cleanPath string) error {
	targetPath, err := utils.IsSafeDeletePath(filepath.Clean(cleanPath))
	if err != nil {
		return fmt.Errorf("Gagal verifikasi path: %w", err)
	}
	if targetPath == "" {
		return fmt.Errorf("File path tidak valid")
	}

	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("File tidak ditemukan")
		}
		return fmt.Errorf("Gagal menghapus file: %w", err)
	}

	return nil
}
