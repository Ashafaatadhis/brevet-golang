package services

import (
	"brevet-api/config"
	"brevet-api/utils"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := uuid.New().String() + ext

	// Buat sub-folder berdasarkan tanggal: uploads/location/2025/07/11
	now := time.Now()
	subPath := filepath.Join(location, now.Format("2006"), now.Format("01"), now.Format("02"))
	saveDir := filepath.Join(s.BaseDir, filepath.Clean(subPath))

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

	publicPath := filepath.ToSlash(filepath.Join(subPath, filename))

	if config.GetEnv("USE_CDN", "false") == "true" {
		cdnBase := config.GetEnv("CDN_URL", "https://cdn.tcugapps.com")
		return fmt.Sprintf("%s/%s", strings.TrimRight(cdnBase, "/"), publicPath), nil
	}

	return fmt.Sprintf("/uploads/%s", publicPath), nil

}

// DeleteFile deletes a file from the server after validating the path
func (s *FileService) DeleteFile(cleanPath string) error {
	// Deteksi jika cleanPath adalah URL (misalnya https://example.com/uploads/...)
	if strings.HasPrefix(cleanPath, "http://") || strings.HasPrefix(cleanPath, "https://") {
		parsed, err := url.Parse(cleanPath)
		if err != nil {
			return fmt.Errorf("URL tidak valid: %w", err)
		}
		// Ambil hanya path lokal dari URL
		cleanPath = parsed.Path // hasil: /uploads/xxx/yyy.pdf
	}

	// Pastikan tidak ada path absolut dari luar
	targetPath, err := utils.IsSafeDeletePath(filepath.Clean(cleanPath))
	if err != nil {
		return fmt.Errorf("Gagal verifikasi path: %w", err)
	}
	if targetPath == "" {
		return fmt.Errorf("File path tidak valid")
	}

	// Hapus file
	if err := os.Remove(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("File tidak ditemukan")
		}
		return fmt.Errorf("Gagal menghapus file: %w", err)
	}

	return nil
}
