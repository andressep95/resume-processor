package repository

import (
	"database/sql"
	"fmt"
	"resume-backend-service/internal/domain"

	"github.com/google/uuid"
)

type ProcessedResumeRepository struct {
	db *sql.DB
}

func NewProcessedResumeRepository(db *sql.DB) *ProcessedResumeRepository {
	return &ProcessedResumeRepository{db: db}
}

// Create crea un nuevo CV procesado (simplificado)
func (r *ProcessedResumeRepository) Create(resume *domain.ProcessedResume) error {
	query := `
		INSERT INTO processed_resumes (request_id, user_id, active_version_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		resume.RequestID,
		resume.UserID,
		resume.ActiveVersionID,
		resume.CreatedAt,
		resume.UpdatedAt,
	).Scan(&resume.ID)

	if err != nil {
		return fmt.Errorf("error al crear CV procesado: %w", err)
	}

	return nil
}

// FindByRequestID busca un CV procesado por su request_id
func (r *ProcessedResumeRepository) FindByRequestID(requestID uuid.UUID) (*domain.ProcessedResume, error) {
	query := `
		SELECT id, request_id, user_id, active_version_id, created_at, updated_at
		FROM processed_resumes
		WHERE request_id = $1
	`

	var resume domain.ProcessedResume
	err := r.db.QueryRow(query, requestID).Scan(
		&resume.ID,
		&resume.RequestID,
		&resume.UserID,
		&resume.ActiveVersionID,
		&resume.CreatedAt,
		&resume.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("CV procesado no encontrado para request_id: %s", requestID)
	}

	if err != nil {
		return nil, fmt.Errorf("error al buscar CV procesado: %w", err)
	}

	return &resume, nil
}

// UpdateActiveVersion actualiza la versión activa de un CV
func (r *ProcessedResumeRepository) UpdateActiveVersion(requestID uuid.UUID, versionID int64) error {
	query := `UPDATE processed_resumes SET active_version_id = $1, updated_at = CURRENT_TIMESTAMP WHERE request_id = $2`
	
	_, err := r.db.Exec(query, versionID, requestID)
	if err != nil {
		return fmt.Errorf("error al actualizar versión activa: %w", err)
	}
	
	return nil
}

// Delete elimina un CV procesado
func (r *ProcessedResumeRepository) Delete(requestID uuid.UUID) error {
	query := `DELETE FROM processed_resumes WHERE request_id = $1`

	result, err := r.db.Exec(query, requestID)
	if err != nil {
		return fmt.Errorf("error al eliminar CV procesado: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al verificar filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("CV procesado no encontrado: %s", requestID)
	}

	return nil
}
