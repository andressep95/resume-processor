package dto

type ResumeProcessorResponseDTO struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"` // UUID de tracking
}
