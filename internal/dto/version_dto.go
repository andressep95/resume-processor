package dto

import "time"

// VersionListResponse representa la respuesta con listado de versiones
type VersionListResponse struct {
	Status   string              `json:"status"`
	Total    int                 `json:"total"`
	Versions []VersionListItem   `json:"versions"`
}

// VersionListItem representa un item resumido de una versión
type VersionListItem struct {
	ID            int64     `json:"id"`
	RequestID     string    `json:"request_id"`
	VersionNumber int       `json:"version_number"`
	VersionName   string    `json:"version_name"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateVersionRequest representa los datos para crear una nueva versión
type CreateVersionRequest struct {
	StructuredData CVProcessedData `json:"structured_data"`
	VersionName    string          `json:"version_name"`
}

// CreateVersionResponse representa la respuesta al crear una nueva versión
type CreateVersionResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	VersionID int64  `json:"version_id"`
}

// VersionDetail representa el detalle completo de una versión específica
type VersionDetail struct {
	Status         string          `json:"status"`
	VersionID      int64           `json:"version_id"`
	VersionNumber  int             `json:"version_number"`
	VersionName    string          `json:"version_name"`
	CreatedBy      string          `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	StructuredData CVProcessedData `json:"structured_data"`
}