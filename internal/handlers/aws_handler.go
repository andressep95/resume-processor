package handlers

import (
	"encoding/json"
	"log"
	"resume-backend-service/internal/domain"
	"resume-backend-service/internal/dto"
	"resume-backend-service/internal/repository"
	"resume-backend-service/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AWSHandler struct {
	resumeRequestRepo   *repository.ResumeRequestRepository
	processedResumeRepo *repository.ProcessedResumeRepository
	resumeVersionRepo   *repository.ResumeVersionRepository
}

func NewAWSHandler(resumeRequestRepo *repository.ResumeRequestRepository, processedResumeRepo *repository.ProcessedResumeRepository, resumeVersionRepo *repository.ResumeVersionRepository) *AWSHandler {
	return &AWSHandler{
		resumeRequestRepo:   resumeRequestRepo,
		processedResumeRepo: processedResumeRepo,
		resumeVersionRepo:   resumeVersionRepo,
	}
}

func (h *AWSHandler) ProcessResumeResultsHandler(c *fiber.Ctx) error {
	// Log del body raw para debug
	bodyRaw := c.Body()
	log.Printf("📥 Body raw recibido (%d bytes): %s", len(bodyRaw), string(bodyRaw))

	// Log de headers para debug
	c.Request().Header.VisitAll(func(key, value []byte) {
		log.Printf("  %s: %s", string(key), string(value))
	})

	// Parsear la respuesta completa de AWS Lambda
	var lambdaResponse dto.AWSLambdaResponse
	if err := c.BodyParser(&lambdaResponse); err != nil {
		log.Printf("❌ Error al parsear el cuerpo de la solicitud: %v", err)

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al parsear el cuerpo de la solicitud.",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// 1. Validar request_id
	if lambdaResponse.RequestID == "" {
		log.Printf("❌ Request ID no encontrado en la respuesta de AWS")
		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Request ID no encontrado en la respuesta.",
		}
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// 2. Parsear request_id como UUID
	requestID, err := uuid.Parse(lambdaResponse.RequestID)
	if err != nil {
		log.Printf("❌ Request ID inválido: %s - Error: %v", lambdaResponse.RequestID, err)
		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Request ID inválido.",
		}
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	log.Printf("📋 Procesando resultado para request_id: %s", requestID)

	// 3. Buscar solicitud original en la base de datos
	resumeRequest, err := h.resumeRequestRepo.FindByRequestID(requestID)
	if err != nil {
		log.Printf("❌ Error al buscar solicitud: %v", err)
		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Solicitud no encontrada.",
		}
		return c.Status(fiber.StatusNotFound).JSON(response)
	}

	log.Printf("✅ Solicitud encontrada: user_id=%s, filename=%s", resumeRequest.UserID, resumeRequest.OriginalFilename)

	// 4. Verificar el status de AWS
	if lambdaResponse.Status != "success" {
		log.Printf("⚠️  AWS reportó status: %s", lambdaResponse.Status)

		// Marcar solicitud como fallida
		if err := h.resumeRequestRepo.MarkAsFailed(requestID, "AWS Lambda reportó status: "+lambdaResponse.Status); err != nil {
			log.Printf("❌ Error al marcar solicitud como fallida: %v", err)
		}

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error en el procesamiento de AWS.",
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// 5. Sanitizar fechas en certificaciones
	structuredDataMap, err := utils.SanitizeStructuredData(lambdaResponse.StructuredData)
	if err != nil {
		log.Printf("❌ Error al sanitizar datos estructurados: %v", err)
		h.resumeRequestRepo.MarkAsFailed(requestID, "Error al sanitizar datos estructurados")

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al sanitizar datos estructurados.",
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// Convertir de vuelta a la estructura tipada
	structuredDataBytes, _ := json.Marshal(structuredDataMap)
	var sanitizedStructuredData dto.CVProcessedData
	if err := json.Unmarshal(structuredDataBytes, &sanitizedStructuredData); err != nil {
		log.Printf("❌ Error al convertir datos sanitizados: %v", err)
		h.resumeRequestRepo.MarkAsFailed(requestID, "Error al convertir datos sanitizados")

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al convertir datos sanitizados.",
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// 6. Log del JSON parseado de forma legible
	jsonPretty, _ := json.MarshalIndent(sanitizedStructuredData, "", "  ")
	log.Printf("✅ CV procesado correctamente:")
	log.Printf("   🆔 Request ID: %s", requestID)
	log.Printf("   👤 User ID: %s", resumeRequest.UserID)
	log.Printf("   📄 Input: %s", lambdaResponse.InputFile)
	log.Printf("   📄 Output: %s", lambdaResponse.OutputFile)
	log.Printf("   ⏱️  Tiempo: %dms", lambdaResponse.ProcessingTimeMs)
	log.Printf("   👤 Nombre: %s", sanitizedStructuredData.Header.Name)
	log.Printf("   📧 Email: %s", sanitizedStructuredData.Header.Contact.Email)
	log.Printf("   📞 Teléfono: %s", sanitizedStructuredData.Header.Contact.Phone)
	log.Printf("   🎓 Educación: %d registros", len(sanitizedStructuredData.Education))
	log.Printf("   💼 Experiencia: %d registros", len(sanitizedStructuredData.ProfessionalExperience))
	log.Printf("   🏆 Certificaciones: %d registros", len(sanitizedStructuredData.Certifications))
	log.Printf("   🚀 Proyectos: %d registros", len(sanitizedStructuredData.Projects))
	log.Printf("   🛠️  Skills: %d registros", len(sanitizedStructuredData.TechnicalSkills.Skills))
	log.Printf("\n📋 Datos completos:\n%s", string(jsonPretty))

	// 7. Crear CV procesado y primera versión
	processedResume := domain.NewProcessedResume(requestID, resumeRequest.UserID)
	if err := h.processedResumeRepo.Create(processedResume); err != nil {
		log.Printf("❌ Error al crear ProcessedResume: %v", err)
		h.resumeRequestRepo.MarkAsFailed(requestID, "Error al crear CV procesado")

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al crear CV procesado.",
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	// 8. Crear primera versión del CV
	versionID, err := h.resumeVersionRepo.CreateVersion(
		requestID,
		resumeRequest.UserID,
		&sanitizedStructuredData,
		"Versión inicial",
		"system",
	)
	if err != nil {
		log.Printf("❌ Error al crear versión inicial: %v", err)
		h.resumeRequestRepo.MarkAsFailed(requestID, "Error al crear versión inicial")

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al crear versión inicial.",
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	log.Printf("✅ CV procesado guardado: resume_id=%d, version_id=%d", processedResume.ID, versionID)

	// 9. Marcar solicitud como completada
	if err := h.resumeRequestRepo.MarkAsCompleted(requestID, lambdaResponse.OutputFile, lambdaResponse.ProcessingTimeMs); err != nil {
		log.Printf("⚠️  Error al marcar solicitud como completada: %v", err)
		// No fallar la operación, solo log
	}

	log.Printf("🎉 Procesamiento completo para request_id: %s", requestID)

	response := dto.AWSProcessResponse{
		Status:  "success",
		Message: "Datos procesados y guardados correctamente.",
	}

	return c.Status(fiber.StatusOK).JSON(response)

}
