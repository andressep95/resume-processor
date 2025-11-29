package dto

type CVProcessorRequestDTO struct {
	Instructions string `json:"instructions"`
	Language     string `json:"language"`
}

type CVProcessorResponseDTO struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
