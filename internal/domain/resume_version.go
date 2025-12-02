package domain

import (
	"encoding/json"
	"resume-backend-service/internal/dto"
	"time"

	"github.com/google/uuid"
)

// ResumeVersion representa una versión específica de un CV
type ResumeVersion struct {
	ID             int64           `json:"id" db:"id"`
	RequestID      uuid.UUID       `json:"request_id" db:"request_id"`
	UserID         string          `json:"user_id" db:"user_id"`
	VersionNumber  int             `json:"version_number" db:"version_number"`
	StructuredData json.RawMessage `json:"structured_data" db:"structured_data"`
	VersionName    string          `json:"version_name" db:"version_name"`
	CreatedBy      string          `json:"created_by" db:"created_by"`
	Status         string          `json:"status" db:"status"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// NewResumeVersion crea una nueva versión de CV
func NewResumeVersion(requestID uuid.UUID, userID string, cvData *dto.CVProcessedData, versionName, createdBy string) (*ResumeVersion, error) {
	structuredDataBytes, err := json.Marshal(cvData)
	if err != nil {
		return nil, err
	}

	return &ResumeVersion{
		RequestID:      requestID,
		UserID:         userID,
		StructuredData: structuredDataBytes,
		VersionName:    versionName,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
	}, nil
}

// GetStructuredData deserializa los datos estructurados
func (rv *ResumeVersion) GetStructuredData() (*dto.CVProcessedData, error) {
	var cvData dto.CVProcessedData
	if err := json.Unmarshal(rv.StructuredData, &cvData); err != nil {
		return nil, err
	}
	return &cvData, nil
}