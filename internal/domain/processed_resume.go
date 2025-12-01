package domain

import (
	"encoding/json"
	"resume-backend-service/internal/dto"
	"time"

	"github.com/google/uuid"
)

// ProcessedResume representa un CV procesado con datos estructurados
type ProcessedResume struct {
	ID                   int64           `json:"id" db:"id"`
	RequestID            uuid.UUID       `json:"request_id" db:"request_id"`
	UserID               string          `json:"user_id" db:"user_id"`
	StructuredData       json.RawMessage `json:"structured_data" db:"structured_data"`
	CVName               string          `json:"cv_name" db:"cv_name"`
	CVEmail              string          `json:"cv_email" db:"cv_email"`
	CVPhone              string          `json:"cv_phone" db:"cv_phone"`
	EducationCount       int             `json:"education_count" db:"education_count"`
	ExperienceCount      int             `json:"experience_count" db:"experience_count"`
	CertificationsCount  int             `json:"certifications_count" db:"certifications_count"`
	ProjectsCount        int             `json:"projects_count" db:"projects_count"`
	SkillsCount          int             `json:"skills_count" db:"skills_count"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

// NewProcessedResume crea un nuevo CV procesado desde los datos de AWS
func NewProcessedResume(requestID uuid.UUID, userID string, cvData *dto.CVProcessedData) (*ProcessedResume, error) {
	// Convertir CVProcessedData a JSON
	structuredDataBytes, err := json.Marshal(cvData)
	if err != nil {
		return nil, err
	}

	return &ProcessedResume{
		RequestID:           requestID,
		UserID:              userID,
		StructuredData:      structuredDataBytes,
		CVName:              cvData.Header.Name,
		CVEmail:             cvData.Header.Contact.Email,
		CVPhone:             cvData.Header.Contact.Phone,
		EducationCount:      len(cvData.Education),
		ExperienceCount:     len(cvData.ProfessionalExperience),
		CertificationsCount: len(cvData.Certifications),
		ProjectsCount:       len(cvData.Projects),
		SkillsCount:         len(cvData.TechnicalSkills.Skills),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}

// GetStructuredData deserializa los datos estructurados a CVProcessedData
func (p *ProcessedResume) GetStructuredData() (*dto.CVProcessedData, error) {
	var cvData dto.CVProcessedData
	if err := json.Unmarshal(p.StructuredData, &cvData); err != nil {
		return nil, err
	}
	return &cvData, nil
}
