package services

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"resume-backend-service/internal/domain"
	"resume-backend-service/internal/dto"
	"resume-backend-service/internal/repository"
	"resume-backend-service/pkg/client"
	"resume-backend-service/pkg/converter"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ResumeService struct {
	presignedURLClient  *client.PresignedURLClient
	resumeRequestRepo   *repository.ResumeRequestRepository
}

func NewResumeService(presignedURLClient *client.PresignedURLClient, resumeRequestRepo *repository.ResumeRequestRepository) *ResumeService {
	return &ResumeService{
		presignedURLClient:  presignedURLClient,
		resumeRequestRepo:   resumeRequestRepo,
	}
}

// Regla de Negocio: Extensiones permitidas (Definidas en el Servicio)
var allowedExtensions = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".docx": true,
	// Nota: .doc (formato antiguo) no está soportado sin LibreOffice
}

func (s *ResumeService) ProcessResume(userID string, instructions string, language string, fileHeader *multipart.FileHeader) (dto.ResumeProcessorResponseDTO, error) {

	// 1. Validación de Formato (Mantenido en el Service)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExtensions[ext] {
		// Retornamos un error de Fiber que el handler puede mapear a 400 Bad Request
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusBadRequest, "Formato de archivo no permitido. Permite: .pdf, .txt, .docx")
	}

	// 2. Crear solicitud de procesamiento con request_id
	resumeRequest := domain.NewResumeRequest(
		userID,
		fileHeader.Filename,
		ext,
		fileHeader.Size,
		language,
		instructions,
	)

	// 3. Guardar solicitud en base de datos (estado: pending)
	if err := s.resumeRequestRepo.Create(resumeRequest); err != nil {
		log.Printf("❌ Error al guardar solicitud: %v", err)
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al procesar solicitud.")
	}

	log.Printf("📝 Solicitud creada: request_id=%s, user_id=%s, filename=%s", resumeRequest.RequestID, userID, fileHeader.Filename)

	// 4. Convertir archivo a PDF (si no lo es ya)
	pdfBytes, pdfFilename, err := converter.ConvertToPDF(fileHeader)
	if err != nil {
		log.Printf("Error al convertir archivo a PDF: %v", err)
		// Marcar como fallida
		s.resumeRequestRepo.MarkAsFailed(resumeRequest.RequestID, "Error al convertir archivo a PDF")
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al procesar el archivo.")
	}

	log.Printf("Archivo convertido a PDF exitosamente: %s (%d bytes)", pdfFilename, len(pdfBytes))

	// 5. Obtener URL firmada del servicio de presigned URLs
	// IMPORTANTE: Se envía request_id para que sea incluido en la firma de la presigned URL
	log.Printf("🔑 Solicitando URL firmada - RequestID: %s, Filename: %s, Language: %s",
		resumeRequest.RequestID, pdfFilename, language)

	presignedResp, err := s.presignedURLClient.GetUploadURL(
		pdfFilename,
		"application/pdf",
		resumeRequest.RequestID.String(),
		language,
		instructions,
	)
	if err != nil {
		log.Printf("❌ Error al obtener URL firmada: %v", err)
		s.resumeRequestRepo.MarkAsFailed(resumeRequest.RequestID, "Error al obtener URL firmada")
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al preparar la subida del archivo.")
	}

	log.Printf("URL firmada obtenida exitosamente (expira en: %s)", presignedResp.ExpiresIn)

	// 6. Subir el PDF a S3 usando la URL firmada con los metadatos (INCLUIR REQUEST_ID)
	if err := s.uploadToS3(presignedResp.URL, pdfBytes, resumeRequest.RequestID.String(), language, instructions); err != nil {
		log.Printf("Error al subir archivo a S3: %v", err)
		s.resumeRequestRepo.MarkAsFailed(resumeRequest.RequestID, "Error al subir archivo a S3")
		return dto.ResumeProcessorResponseDTO{}, fiber.NewError(fiber.StatusInternalServerError, "Error al subir el archivo.")
	}

	log.Printf("Archivo subido exitosamente a S3: %s", pdfFilename)

	// 7. Marcar solicitud como subida (estado: uploaded)
	// La URL de S3 se puede extraer del presignedResp.URL (quitar query params)
	s3InputURL := strings.Split(presignedResp.URL, "?")[0]
	if err := s.resumeRequestRepo.MarkAsUploaded(resumeRequest.RequestID, s3InputURL); err != nil {
		log.Printf("⚠️  Error al actualizar estado de solicitud: %v", err)
		// No fallar la operación, solo log
	}

	// 8. Retorno de DTO de éxito CON REQUEST_ID
	return dto.ResumeProcessorResponseDTO{
		Status:    "accepted",
		Message:   "Solicitud encolada para procesamiento.",
		RequestID: resumeRequest.RequestID.String(),
	}, nil
}

// uploadToS3 sube un archivo a S3 usando una URL firmada
// Los headers de metadata DEBEN coincidir exactamente con los usados al generar la presigned URL
func (s *ResumeService) uploadToS3(presignedURL string, fileData []byte, requestID, language, instructions string) error {
	req, err := http.NewRequest("PUT", presignedURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("error al crear request de subida: %w", err)
	}

	// Headers requeridos - DEBEN coincidir con los metadatos de la presigned URL
	req.Header.Set("Content-Type", "application/pdf")
	req.Header.Set("x-amz-meta-request-id", requestID)      // Request ID para tracking
	req.Header.Set("x-amz-meta-language", language)
	req.Header.Set("x-amz-meta-instructions", instructions)

	log.Printf("🔄 Subiendo a S3 - RequestID: %s, Size: %d bytes, Language: %s",
		requestID, len(fileData), language)

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
