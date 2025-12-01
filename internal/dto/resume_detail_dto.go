package dto

import (
	"time"
)

// ResumeDetailDTO representa el detalle completo de un CV procesado
type ResumeDetailDTO struct {
	// Datos de la solicitud
	RequestID        string    `json:"request_id"`
	OriginalFilename string    `json:"original_filename"`
	OriginalFileType string    `json:"original_file_type"`
	FileSizeBytes    int64     `json:"file_size_bytes"`
	Language         string    `json:"language"`
	Instructions     string    `json:"instructions,omitempty"`
	Status           string    `json:"status"`
	S3InputURL       string    `json:"s3_input_url,omitempty"`
	S3OutputURL      string    `json:"s3_output_url,omitempty"`
	ProcessingTimeMs int64     `json:"processing_time_ms,omitempty"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UploadedAt       *time.Time `json:"uploaded_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	
	// Datos del CV procesado (si existe)
	StructuredData *CVProcessedData `json:"structured_data,omitempty"`
}
