package repository

import (
	"database/sql"
	"fmt"
)

type ResumeListItem struct {
	RequestID        string
	OriginalFilename string
	Status           string
	CreatedAt        string
	CompletedAt      sql.NullString
	FullName         sql.NullString
	Email            sql.NullString
}

// GetUserResumes obtiene el listado resumido de CVs de un usuario
func (r *ResumeRequestRepository) GetUserResumes(userID string) ([]ResumeListItem, error) {
	query := `
		SELECT 
			rr.request_id,
			rr.original_filename,
			rr.status,
			rr.created_at,
			rr.completed_at,
			pr.full_name,
			pr.email
		FROM resume_requests rr
		LEFT JOIN processed_resumes pr ON rr.request_id = pr.request_id
		WHERE rr.user_id = $1
		ORDER BY rr.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener listado de CVs: %w", err)
	}
	defer rows.Close()

	var items []ResumeListItem
	for rows.Next() {
		var item ResumeListItem
		err := rows.Scan(
			&item.RequestID,
			&item.OriginalFilename,
			&item.Status,
			&item.CreatedAt,
			&item.CompletedAt,
			&item.FullName,
			&item.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
