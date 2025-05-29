package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// IsSafeDeletePath is function to check if path is safe to delete
func IsSafeDeletePath(cleanPath string) (string, error) {
	uploadRoot := os.Getenv("UPLOAD_DIR")
	if uploadRoot == "" {
		uploadRoot = "" // fallback default
	}

	absUploadRoot, err := filepath.Abs(uploadRoot)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(uploadRoot, cleanPath)
	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(absTargetPath, absUploadRoot) {
		return "", nil // bukan path valid
	}

	return absTargetPath, nil
}

// IsSafePath is a function to check if a path is safe
func IsSafePath(baseDir, targetPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// Pastikan absTarget diawali dengan absBase (path traversal dicegah)
	return strings.HasPrefix(absTarget, absBase)
}
