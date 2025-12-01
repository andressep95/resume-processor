package dto

// PresignedURLRequest es la petición que se envía al servicio de presigned URLs
type PresignedURLRequest struct {
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Metadata    PresignedMetadata `json:"metadata"`
}

// PresignedMetadata contiene los metadatos personalizados para el archivo
type PresignedMetadata struct {
	Language     string `json:"language"`
	Instructions string `json:"instructions"`
	UserEmail    string `json:"user_email"`
}

// PresignedURLResponse es la respuesta del servicio de presigned URLs
type PresignedURLResponse struct {
	URL       string `json:"url"`
	ExpiresIn string `json:"expires_in"`
}
