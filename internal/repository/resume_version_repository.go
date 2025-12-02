package repository

import (
	"database/sql"
	"encoding/json"
	"resume-backend-service/internal/domain"
	"resume-backend-service/internal/dto"

	"github.com/google/uuid"
)

type ResumeVersionRepository struct {
	db *sql.DB
}

func NewResumeVersionRepository(db *sql.DB) *ResumeVersionRepository {
	return &ResumeVersionRepository{db: db}
}

// CreateVersion crea una nueva versión usando la función SQL
func (r *ResumeVersionRepository) CreateVersion(requestID uuid.UUID, userID string, cvData *dto.CVProcessedData, versionName, createdBy string) (int64, error) {
	structuredDataBytes, err := json.Marshal(cvData)
	if err != nil {
		return 0, err
	}

	var versionID int64
	query := `SELECT create_resume_version($1, $2, $3, $4, $5)`
	
	err = r.db.QueryRow(query, requestID, userID, structuredDataBytes, 
		sql.NullString{String: versionName, Valid: versionName != ""}, createdBy).Scan(&versionID)
	
	return versionID, err
}

// ActivateVersion activa una versión específica
func (r *ResumeVersionRepository) ActivateVersion(requestID uuid.UUID, versionID int64) error {
	query := `SELECT activate_resume_version($1, $2)`
	var success bool
	
	err := r.db.QueryRow(query, requestID, versionID).Scan(&success)
	if err != nil {
		return err
	}
	
	if !success {
		return sql.ErrNoRows
	}
	
	return nil
}

// GetVersionsByRequestID obtiene todas las versiones activas de un CV
func (r *ResumeVersionRepository) GetVersionsByRequestID(requestID uuid.UUID) ([]*domain.ResumeVersion, error) {
	query := `
		SELECT id, request_id, user_id, version_number, structured_data, 
		       COALESCE(version_name, ''), created_by, status, created_at
		FROM resume_versions 
		WHERE request_id = $1 AND status = 'active'
		ORDER BY version_number DESC`
	
	rows, err := r.db.Query(query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.ResumeVersion
	for rows.Next() {
		version := &domain.ResumeVersion{}
		err := rows.Scan(
			&version.ID,
			&version.RequestID,
			&version.UserID,
			&version.VersionNumber,
			&version.StructuredData,
			&version.VersionName,
			&version.CreatedBy,
			&version.Status,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, rows.Err()
}

// GetVersionByID obtiene una versión específica por ID (solo activas)
func (r *ResumeVersionRepository) GetVersionByID(versionID int64) (*domain.ResumeVersion, error) {
	query := `
		SELECT id, request_id, user_id, version_number, structured_data, 
		       COALESCE(version_name, ''), created_by, status, created_at
		FROM resume_versions 
		WHERE id = $1 AND status = 'active'`
	
	version := &domain.ResumeVersion{}
	err := r.db.QueryRow(query, versionID).Scan(
		&version.ID,
		&version.RequestID,
		&version.UserID,
		&version.VersionNumber,
		&version.StructuredData,
		&version.VersionName,
		&version.CreatedBy,
		&version.Status,
		&version.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return version, nil
}

// SoftDeleteVersion marca una versión como eliminada
func (r *ResumeVersionRepository) SoftDeleteVersion(versionID int64, userID string) error {
	query := `SELECT soft_delete_resume_version($1, $2)`
	var success bool
	
	err := r.db.QueryRow(query, versionID, userID).Scan(&success)
	if err != nil {
		return err
	}
	
	if !success {
		return sql.ErrNoRows
	}
	
	return nil
}