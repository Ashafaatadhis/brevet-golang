package utils

import (
	"path/filepath"
	"slices"
	"strings"
)

// AllowedImageExtensions is a list of allowed image extensions
var AllowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

// AllowedDocumentExtensions is a list of allowed document extensions
var AllowedDocumentExtensions = []string{".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx", ".txt"}

// IsAllowedExtension checks if the file extension is allowed
func IsAllowedExtension(filename string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(allowedExts, ext)
}
