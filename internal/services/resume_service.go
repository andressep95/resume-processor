package services

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"resume-backend-service/internal/dto"
	"resume-backend-service/pkg/client"
	"resume-backend-service/pkg/converter"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ResumeService struct {
	presignedURLClient *client.PresignedURLClient
}

func NewResumeService(presignedURLClient *client.PresignedURLClient) *ResumeService {
	return &ResumeService{
		presignedURLClient: presignedURLClient,
	}
}

// Regla de Negocio: Extensiones permitidas (Definidas en el Servicio)
var allowedExtensions = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".docx": true,
	// Nota: .doc (formato antiguo) no está soportado sin LibreOffice
}

func (s *ResumeService) ProcessResume(instructions string, language string, userEmail string, fileHeader *multipart.FileHeader) (dto.ResumeProcessorResponseDTO, error) {

	// 1. Validación de Formato (Mantenido en el Service)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExtensions[ext] {
		// Retornamos un error de Fiber que el handler puede mapear a 400 Bad Request
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusBadRequest, "Formato de archivo no permitido. Permite: .pdf, .txt, .docx")
	}

	// 2. Convertir archivo a PDF (si no lo es ya)
	pdfBytes, pdfFilename, err := converter.ConvertToPDF(fileHeader)
	if err != nil {
		log.Printf("Error al convertir archivo a PDF: %v", err)
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al procesar el archivo.")
	}

	log.Printf("Archivo convertido a PDF exitosamente: %s (%d bytes)", pdfFilename, len(pdfBytes))

	// 3. Obtener URL firmada del servicio de presigned URLs
	// TEMPORAL: No enviar userEmail hasta que el servicio de presigned URLs lo soporte
	log.Printf("🔑 Solicitando URL firmada - Filename: %s, Language: %s, UserEmail: %s (no enviado aún)", pdfFilename, language, userEmail)
	presignedResp, err := s.presignedURLClient.GetUploadURL(
		pdfFilename,
		"application/pdf",
		language,
		instructions,
		"", // userEmail temporalmente vacío hasta que el servicio lo soporte
	)
	if err != nil {
		log.Printf("❌ Error al obtener URL firmada: %v", err)
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al preparar la subida del archivo.")
	}

	log.Printf("URL firmada obtenida exitosamente (expira en: %s)", presignedResp.ExpiresIn)

	// 4. Subir el PDF a S3 usando la URL firmada (metadata ya incluida en la URL)
	if err := s.uploadToS3(presignedResp.URL, pdfBytes); err != nil {
		log.Printf("Error al subir archivo a S3: %v", err)
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al subir el archivo.")
	}

	log.Printf("Archivo subido exitosamente a S3: %s", pdfFilename)

	// 5. Retorno de DTO de éxito
	return dto.ResumeProcessorResponseDTO{
		Status:  "accepted",
		Message: "Solicitud encolada para procesamiento.",
	}, nil
}

// uploadToS3 sube un archivo a S3 usando una URL firmada
// Los metadatos ya están incluidos en la URL firmada por el servicio de presigned URLs
func (s *ResumeService) uploadToS3(presignedURL string, fileData []byte) error {
	req, err := http.NewRequest("PUT", presignedURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("error al crear request de subida: %w", err)
	}

	// Headers requeridos
	// NOTA: Los metadatos ya están incluidos en la URL firmada por el servicio de presigned URLs
	// No se deben enviar manualmente o la firma será inválida (403)
	req.Header.Set("Content-Type", "application/pdf")

	log.Printf("🔄 Subiendo a S3 - Size: %d bytes, Content-Type: %s", len(fileData), req.Header.Get("Content-Type"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error al ejecutar subida: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("❌ S3 Response Status: %d, Headers: %v", resp.StatusCode, resp.Header)
		return fmt.Errorf("error al subir archivo a S3 (status %d)", resp.StatusCode)
	}

	log.Printf("✅ S3 Response Status: %d", resp.StatusCode)
	return nil
}
