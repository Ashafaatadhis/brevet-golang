package dto

// UploadRequest represents the request body for file upload
type UploadRequest struct {
	Location string `form:"location" validate:"required"`
}

// DeleteRequest represents the request body for file deletion
type DeleteRequest struct {
	FilePath string `json:"file_path" validate:"required"`
}
