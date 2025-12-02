package handlers

import (
	"database/sql"
	"encoding/json"
	"resume-backend-service/internal/dto"
	"resume-backend-service/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ResumeVersionHandler struct {
	resumeVersionRepo   *repository.ResumeVersionRepository
	processedResumeRepo *repository.ProcessedResumeRepository
}

func NewResumeVersionHandler(resumeVersionRepo *repository.ResumeVersionRepository, processedResumeRepo *repository.ProcessedResumeRepository) *ResumeVersionHandler {
	return &ResumeVersionHandler{
		resumeVersionRepo:   resumeVersionRepo,
		processedResumeRepo: processedResumeRepo,
	}
}

// GetVersions obtiene todas las versiones de un CV
func (h *ResumeVersionHandler) GetVersions(c *fiber.Ctx) error {
	requestIDStr := c.Params("request_id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Request ID inválido",
		})
	}

	userID := c.Locals("user_subject").(string)

	// Verificar que el CV pertenece al usuario
	processedResume, err := h.processedResumeRepo.FindByRequestID(requestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "CV no encontrado",
		})
	}

	if processedResume.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "No tienes acceso a este CV",
		})
	}

	versions, err := h.resumeVersionRepo.GetVersionsByRequestID(requestID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al obtener versiones",
		})
	}

	// Convertir a DTO
	versionItems := make([]dto.VersionListItem, len(versions))
	for i, v := range versions {
		versionItems[i] = dto.VersionListItem{
			ID:            v.ID,
			RequestID:     v.RequestID.String(),
			VersionNumber: v.VersionNumber,
			VersionName:   v.VersionName,
			CreatedBy:     v.CreatedBy,
			CreatedAt:     v.CreatedAt,
		}
	}

	response := dto.VersionListResponse{
		Status:   "success",
		Total:    len(versions),
		Versions: versionItems,
	}

	return c.JSON(response)
}

// CreateVersion crea una nueva versión del CV
func (h *ResumeVersionHandler) CreateVersion(c *fiber.Ctx) error {
	requestIDStr := c.Params("request_id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Request ID inválido",
		})
	}

	userID := c.Locals("user_subject").(string)

	// Verificar que el CV pertenece al usuario
	processedResume, err := h.processedResumeRepo.FindByRequestID(requestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "CV no encontrado",
		})
	}

	if processedResume.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "No tienes acceso a este CV",
		})
	}

	// Parsear el body
	var req dto.CreateVersionRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Datos inválidos",
		})
	}

	// Crear nueva versión
	versionID, err := h.resumeVersionRepo.CreateVersion(
		requestID,
		userID,
		&req.StructuredData,
		req.VersionName,
		"user",
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al crear versión",
		})
	}

	response := dto.CreateVersionResponse{
		Status:    "success",
		Message:   "Versión creada correctamente",
		VersionID: versionID,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// ActivateVersion activa una versión específica
func (h *ResumeVersionHandler) ActivateVersion(c *fiber.Ctx) error {
	requestIDStr := c.Params("request_id")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Request ID inválido",
		})
	}

	versionIDStr := c.Params("version_id")
	versionID, err := strconv.ParseInt(versionIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Version ID inválido",
		})
	}

	userID := c.Locals("user_subject").(string)

	// Verificar que el CV pertenece al usuario
	processedResume, err := h.processedResumeRepo.FindByRequestID(requestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "CV no encontrado",
		})
	}

	if processedResume.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "No tienes acceso a este CV",
		})
	}

	// Activar versión
	err = h.resumeVersionRepo.ActivateVersion(requestID, versionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Versión no encontrada",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Versión activada correctamente",
	})
}

// GetVersionDetail obtiene el detalle de una versión específica
func (h *ResumeVersionHandler) GetVersionDetail(c *fiber.Ctx) error {
	versionIDStr := c.Params("version_id")
	versionID, err := strconv.ParseInt(versionIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Version ID inválido",
		})
	}

	userID := c.Locals("user_subject").(string)

	// Obtener versión
	version, err := h.resumeVersionRepo.GetVersionByID(versionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Versión no encontrada",
		})
	}

	// Verificar que pertenece al usuario
	if version.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "No tienes acceso a esta versión",
		})
	}

	// Deserializar datos estructurados
	var structuredData dto.CVProcessedData
	if err := json.Unmarshal(version.StructuredData, &structuredData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al procesar datos",
		})
	}

	response := dto.VersionDetail{
		Status:         "success",
		VersionID:      version.ID,
		VersionNumber:  version.VersionNumber,
		VersionName:    version.VersionName,
		CreatedBy:      version.CreatedBy,
		CreatedAt:      version.CreatedAt,
		StructuredData: structuredData,
	}

	return c.JSON(response)
}

// DeleteVersion elimina (soft delete) una versión específica
func (h *ResumeVersionHandler) DeleteVersion(c *fiber.Ctx) error {
	versionIDStr := c.Params("version_id")
	versionID, err := strconv.ParseInt(versionIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Version ID inválido",
		})
	}

	userID := c.Locals("user_subject").(string)

	// Eliminar versión (soft delete)
	err = h.resumeVersionRepo.SoftDeleteVersion(versionID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "Versión no encontrada o no se puede eliminar",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error al eliminar versión",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Versión eliminada correctamente",
	})
}