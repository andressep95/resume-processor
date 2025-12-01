package handlers

import (
	"log"
	"resume-backend-service/internal/dto"
	"resume-backend-service/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ResumeListHandler struct {
	resumeRequestRepo   *repository.ResumeRequestRepository
	processedResumeRepo *repository.ProcessedResumeRepository
}

func NewResumeListHandler(resumeRequestRepo *repository.ResumeRequestRepository, processedResumeRepo *repository.ProcessedResumeRepository) *ResumeListHandler {
	return &ResumeListHandler{
		resumeRequestRepo:   resumeRequestRepo,
		processedResumeRepo: processedResumeRepo,
	}
}

// GetMyResumes obtiene el listado de CVs del usuario autenticado
func (h *ResumeListHandler) GetMyResumes(c *fiber.Ctx) error {
	// Extraer user_id del token JWT
	userID := ""
	if subject := c.Locals("user_subject"); subject != nil {
		userID = subject.(string)
	}

	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Usuario no autenticado",
		})
	}

	// Obtener listado de CVs
	items, err := h.resumeRequestRepo.GetUserResumes(userID)
	if err != nil {
		log.Printf("❌ Error al obtener listado de CVs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error al obtener listado de CVs",
		})
	}

	// Convertir a DTO
	resumes := make([]dto.ResumeListItemDTO, 0, len(items))
	for _, item := range items {
		createdAt, _ := time.Parse(time.RFC3339, item.CreatedAt)
		
		resumeDTO := dto.ResumeListItemDTO{
			RequestID:        item.RequestID,
			OriginalFilename: item.OriginalFilename,
			Status:           item.Status,
			CreatedAt:        createdAt,
		}

		if item.CompletedAt.Valid {
			completedAt, _ := time.Parse(time.RFC3339, item.CompletedAt.String)
			resumeDTO.CompletedAt = &completedAt
		}

		if item.FullName.Valid {
			resumeDTO.FullName = item.FullName.String
		}

		if item.Email.Valid {
			resumeDTO.Email = item.Email.String
		}

		resumes = append(resumes, resumeDTO)
	}

	response := dto.ResumeListResponseDTO{
		Total:   len(resumes),
		Resumes: resumes,
	}

	return c.JSON(response)
}

// GetResumeDetail obtiene el detalle completo de un CV por request_id
func (h *ResumeListHandler) GetResumeDetail(c *fiber.Ctx) error {
	// Extraer user_id del token JWT
	userID := ""
	if subject := c.Locals("user_subject"); subject != nil {
		userID = subject.(string)
	}

	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Usuario no autenticado",
		})
	}

	// Obtener request_id del path
	requestIDStr := c.Params("request_id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Request ID inválido",
		})
	}

	// Buscar solicitud
	request, err := h.resumeRequestRepo.FindByRequestID(requestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "CV no encontrado",
		})
	}

	// Verificar que el CV pertenece al usuario
	if request.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "No tienes permiso para ver este CV",
		})
	}

	// Construir respuesta base
	detail := dto.ResumeDetailDTO{
		RequestID:        request.RequestID.String(),
		OriginalFilename: request.OriginalFilename,
		OriginalFileType: request.OriginalFileType,
		FileSizeBytes:    request.FileSizeBytes,
		Language:         request.Language,
		Instructions:     request.Instructions,
		Status:           string(request.Status),
		S3InputURL:       request.S3InputURL,
		S3OutputURL:      request.S3OutputURL,
		ProcessingTimeMs: request.ProcessingTimeMs,
		ErrorMessage:     request.ErrorMessage,
		CreatedAt:        request.CreatedAt,
		UploadedAt:       request.UploadedAt,
		CompletedAt:      request.CompletedAt,
	}

	// Si está completado, obtener datos estructurados
	if request.Status == "completed" {
		processedResume, err := h.processedResumeRepo.FindByRequestID(requestID)
		if err == nil && processedResume != nil {
			structuredData, err := processedResume.GetStructuredData()
			if err == nil {
				detail.StructuredData = structuredData
			}
		}
	}

	return c.JSON(detail)
}
