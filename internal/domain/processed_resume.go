package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProcessedResume representa un CV procesado (referencia a versión activa)
type ProcessedResume struct {
	ID              int64     `json:"id" db:"id"`
	RequestID       uuid.UUID `json:"request_id" db:"request_id"`
	UserID          string    `json:"user_id" db:"user_id"`
	ActiveVersionID *int64    `json:"active_version_id" db:"active_version_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// NewProcessedResume crea un nuevo CV procesado (sin versión activa inicialmente)
func NewProcessedResume(requestID uuid.UUID, userID string) *ProcessedResume {
	return &ProcessedResume{
		RequestID: requestID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SetActiveVersion establece la versión activa
func (p *ProcessedResume) SetActiveVersion(versionID int64) {
	p.ActiveVersionID = &versionID
	p.UpdatedAt = time.Now()
}
