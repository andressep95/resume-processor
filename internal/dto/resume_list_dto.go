package dto

import "time"

// ResumeListItemDTO representa un item en el listado de CVs procesados
type ResumeListItemDTO struct {
	RequestID        string    `json:"request_id"`
	OriginalFilename string    `json:"original_filename"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	FullName         string    `json:"full_name,omitempty"`
	Email            string    `json:"email,omitempty"`
}

// ResumeListResponseDTO representa la respuesta del listado
type ResumeListResponseDTO struct {
	Total   int                  `json:"total"`
	Resumes []ResumeListItemDTO `json:"resumes"`
}
