package repository

import (
	"database/sql"
	"fmt"
	"resume-backend-service/internal/domain"

	"github.com/google/uuid"
)

type ResumeRequestRepository struct {
	db *sql.DB
}

func NewResumeRequestRepository(db *sql.DB) *ResumeRequestRepository {
	return &ResumeRequestRepository{db: db}
}

// Create crea una nueva solicitud de procesamiento
func (r *ResumeRequestRepository) Create(request *domain.ResumeRequest) error {
	query := `
		INSERT INTO resume_requests (
			request_id, user_id, original_filename, original_file_type,
			file_size_bytes, language, instructions, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(
		query,
		request.RequestID,
		request.UserID,
		request.OriginalFilename,
		request.OriginalFileType,
		request.FileSizeBytes,
		request.Language,
		request.Instructions,
		request.Status,
		request.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("error al crear solicitud: %w", err)
	}

	return nil
}

// FindByRequestID busca una solicitud por su request_id
func (r *ResumeRequestRepository) FindByRequestID(requestID uuid.UUID) (*domain.ResumeRequest, error) {
	query := `
		SELECT request_id, user_id, original_filename, original_file_type,
		       file_size_bytes, language, instructions, s3_input_url, s3_output_url,
		       status, processing_time_ms, error_message, created_at, uploaded_at, completed_at
		FROM resume_requests
		WHERE request_id = $1
	`

	var request domain.ResumeRequest
	err := r.db.QueryRow(query, requestID).Scan(
		&request.RequestID,
		&request.UserID,
		&request.OriginalFilename,
		&request.OriginalFileType,
		&request.FileSizeBytes,
		&request.Language,
		&request.Instructions,
		&request.S3InputURL,
		&request.S3OutputURL,
		&request.Status,
		&request.ProcessingTimeMs,
		&request.ErrorMessage,
		&request.CreatedAt,
		&request.UploadedAt,
		&request.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("solicitud no encontrada: %s", requestID)
	}

	if err != nil {
		return nil, fmt.Errorf("error al buscar solicitud: %w", err)
	}

	return &request, nil
}

// FindByUserID busca todas las solicitudes de un usuario
func (r *ResumeRequestRepository) FindByUserID(userID string) ([]*domain.ResumeRequest, error) {
	query := `
		SELECT request_id, user_id, original_filename, original_file_type,
		       file_size_bytes, language, instructions, s3_input_url, s3_output_url,
		       status, processing_time_ms, error_message, created_at, uploaded_at, completed_at
		FROM resume_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al buscar solicitudes del usuario: %w", err)
	}
	defer rows.Close()

	var requests []*domain.ResumeRequest
	for rows.Next() {
		var request domain.ResumeRequest
		err := rows.Scan(
			&request.RequestID,
			&request.UserID,
			&request.OriginalFilename,
			&request.OriginalFileType,
			&request.FileSizeBytes,
			&request.Language,
			&request.Instructions,
			&request.S3InputURL,
			&request.S3OutputURL,
			&request.Status,
			&request.ProcessingTimeMs,
			&request.ErrorMessage,
			&request.CreatedAt,
			&request.UploadedAt,
			&request.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear solicitud: %w", err)
		}
		requests = append(requests, &request)
	}

	return requests, nil
}

// UpdateStatus actualiza el estado de una solicitud
func (r *ResumeRequestRepository) UpdateStatus(requestID uuid.UUID, status domain.ResumeRequestStatus) error {
	query := `UPDATE resume_requests SET status = $1 WHERE request_id = $2`

	result, err := r.db.Exec(query, status, requestID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("solicitud no encontrada: %s", requestID)
	}

	return nil
}

// MarkAsUploaded marca la solicitud como subida a S3
func (r *ResumeRequestRepository) MarkAsUploaded(requestID uuid.UUID, s3InputURL string) error {
	query := `
		UPDATE resume_requests
		SET status = $1, s3_input_url = $2, uploaded_at = NOW()
		WHERE request_id = $3
	`

	result, err := r.db.Exec(query, domain.StatusUploaded, s3InputURL, requestID)
	if err != nil {
		return fmt.Errorf("error al marcar como subido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("solicitud no encontrada: %s", requestID)
	}

	return nil
}

// MarkAsCompleted marca la solicitud como completada
func (r *ResumeRequestRepository) MarkAsCompleted(requestID uuid.UUID, s3OutputURL string, processingTimeMs int64) error {
	query := `
		UPDATE resume_requests
		SET status = $1, s3_output_url = $2, processing_time_ms = $3, completed_at = NOW()
		WHERE request_id = $4
	`

	result, err := r.db.Exec(query, domain.StatusCompleted, s3OutputURL, processingTimeMs, requestID)
	if err != nil {
		return fmt.Errorf("error al marcar como completado: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("solicitud no encontrada: %s", requestID)
	}

	return nil
}

// MarkAsFailed marca la solicitud como fallida
func (r *ResumeRequestRepository) MarkAsFailed(requestID uuid.UUID, errorMessage string) error {
	query := `
		UPDATE resume_requests
		SET status = $1, error_message = $2, completed_at = NOW()
		WHERE request_id = $3
	`

	result, err := r.db.Exec(query, domain.StatusFailed, errorMessage, requestID)
	if err != nil {
		return fmt.Errorf("error al marcar como fallido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("solicitud no encontrada: %s", requestID)
	}

	return nil
}
