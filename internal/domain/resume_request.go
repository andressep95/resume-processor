package domain

import (
	"time"

	"github.com/google/uuid"
)

// ResumeRequestStatus representa los estados posibles de una solicitud
type ResumeRequestStatus string

const (
	StatusPending    ResumeRequestStatus = "pending"
	StatusUploaded   ResumeRequestStatus = "uploaded"
	StatusProcessing ResumeRequestStatus = "processing"
	StatusCompleted  ResumeRequestStatus = "completed"
	StatusFailed     ResumeRequestStatus = "failed"
)

// ResumeRequest representa una solicitud de procesamiento de CV
type ResumeRequest struct {
	RequestID        uuid.UUID           `json:"request_id" db:"request_id"`
	UserID           string              `json:"user_id" db:"user_id"`
	OriginalFilename string              `json:"original_filename" db:"original_filename"`
	OriginalFileType string              `json:"original_file_type" db:"original_file_type"`
	FileSizeBytes    int64               `json:"file_size_bytes" db:"file_size_bytes"`
	Language         string              `json:"language" db:"language"`
	Instructions     string              `json:"instructions" db:"instructions"`
	S3InputURL       string              `json:"s3_input_url,omitempty" db:"s3_input_url"`
	S3OutputURL      string              `json:"s3_output_url,omitempty" db:"s3_output_url"`
	Status           ResumeRequestStatus `json:"status" db:"status"`
	ProcessingTimeMs int64               `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	ErrorMessage     string              `json:"error_message,omitempty" db:"error_message"`
	CreatedAt        time.Time           `json:"created_at" db:"created_at"`
	UploadedAt       *time.Time          `json:"uploaded_at,omitempty" db:"uploaded_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty" db:"completed_at"`
}

// NewResumeRequest crea una nueva solicitud de procesamiento
func NewResumeRequest(userID, filename, fileType string, fileSize int64, language, instructions string) *ResumeRequest {
	return &ResumeRequest{
		RequestID:        uuid.New(),
		UserID:           userID,
		OriginalFilename: filename,
		OriginalFileType: fileType,
		FileSizeBytes:    fileSize,
		Language:         language,
		Instructions:     instructions,
		Status:           StatusPending,
		CreatedAt:        time.Now(),
	}
}

// MarkAsUploaded marca la solicitud como subida a S3
func (r *ResumeRequest) MarkAsUploaded(s3InputURL string) {
	r.Status = StatusUploaded
	r.S3InputURL = s3InputURL
	now := time.Now()
	r.UploadedAt = &now
}

// MarkAsProcessing marca la solicitud como en procesamiento
func (r *ResumeRequest) MarkAsProcessing() {
	r.Status = StatusProcessing
}

// MarkAsCompleted marca la solicitud como completada
func (r *ResumeRequest) MarkAsCompleted(s3OutputURL string, processingTimeMs int64) {
	r.Status = StatusCompleted
	r.S3OutputURL = s3OutputURL
	r.ProcessingTimeMs = processingTimeMs
	now := time.Now()
	r.CompletedAt = &now
}

// MarkAsFailed marca la solicitud como fallida
func (r *ResumeRequest) MarkAsFailed(errorMsg string) {
	r.Status = StatusFailed
	r.ErrorMessage = errorMsg
	now := time.Now()
	r.CompletedAt = &now
}
