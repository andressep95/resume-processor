package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"resume-backend-service/internal/dto"
	"time"
)

// PresignedURLClient maneja las llamadas al servicio de presigned URLs
type PresignedURLClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPresignedURLClient crea una nueva instancia del cliente
func NewPresignedURLClient(baseURL string) *PresignedURLClient {
	return &PresignedURLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetUploadURL obtiene una URL firmada para subir un archivo a S3
func (c *PresignedURLClient) GetUploadURL(filename, contentType, requestID, language, instructions string) (*dto.PresignedURLResponse, error) {
	// Construir el request
	requestBody := dto.PresignedURLRequest{
		Filename:    filename,
		ContentType: contentType,
		Metadata: dto.PresignedMetadata{
			RequestID:    requestID,
			Language:     language,
			Instructions: instructions,
		},
	}

	// Serializar a JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error al serializar request: %w", err)
	}

	// Crear la petición HTTP
	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error al crear request HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Ejecutar la petición
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al ejecutar petición: %w", err)
	}
	defer resp.Body.Close()

	// Leer el body de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta: %w", err)
	}

	// Validar código de estado
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error del servicio de presigned URLs (status %d): %s", resp.StatusCode, string(body))
	}

	// Deserializar respuesta
	var presignedResponse dto.PresignedURLResponse
	if err := json.Unmarshal(body, &presignedResponse); err != nil {
		return nil, fmt.Errorf("error al deserializar respuesta: %w", err)
	}

	return &presignedResponse, nil
}
