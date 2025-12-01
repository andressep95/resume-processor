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

// Create crea un nuevo CV procesado
func (r *ProcessedResumeRepository) Create(resume *domain.ProcessedResume) error {
	query := `
		INSERT INTO processed_resumes (
			request_id, user_id, structured_data, cv_name, cv_email, cv_phone,
			education_count, experience_count, certifications_count, projects_count, skills_count,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		resume.RequestID,
		resume.UserID,
		resume.StructuredData,
		resume.CVName,
		resume.CVEmail,
		resume.CVPhone,
		resume.EducationCount,
		resume.ExperienceCount,
		resume.CertificationsCount,
		resume.ProjectsCount,
		resume.SkillsCount,
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
		SELECT id, request_id, user_id, structured_data, cv_name, cv_email, cv_phone,
		       education_count, experience_count, certifications_count, projects_count, skills_count,
		       created_at, updated_at
		FROM processed_resumes
		WHERE request_id = $1
	`

	var resume domain.ProcessedResume
	err := r.db.QueryRow(query, requestID).Scan(
		&resume.ID,
		&resume.RequestID,
		&resume.UserID,
		&resume.StructuredData,
		&resume.CVName,
		&resume.CVEmail,
		&resume.CVPhone,
		&resume.EducationCount,
		&resume.ExperienceCount,
		&resume.CertificationsCount,
		&resume.ProjectsCount,
		&resume.SkillsCount,
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

// FindByUserID busca todos los CVs procesados de un usuario
func (r *ProcessedResumeRepository) FindByUserID(userID string) ([]*domain.ProcessedResume, error) {
	query := `
		SELECT id, request_id, user_id, structured_data, cv_name, cv_email, cv_phone,
		       education_count, experience_count, certifications_count, projects_count, skills_count,
		       created_at, updated_at
		FROM processed_resumes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al buscar CVs del usuario: %w", err)
	}
	defer rows.Close()

	var resumes []*domain.ProcessedResume
	for rows.Next() {
		var resume domain.ProcessedResume
		err := rows.Scan(
			&resume.ID,
			&resume.RequestID,
			&resume.UserID,
			&resume.StructuredData,
			&resume.CVName,
			&resume.CVEmail,
			&resume.CVPhone,
			&resume.EducationCount,
			&resume.ExperienceCount,
			&resume.CertificationsCount,
			&resume.ProjectsCount,
			&resume.SkillsCount,
			&resume.CreatedAt,
			&resume.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear CV: %w", err)
		}
		resumes = append(resumes, &resume)
	}

	return resumes, nil
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
