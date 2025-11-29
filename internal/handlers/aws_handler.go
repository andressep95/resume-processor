package handlers

import (
	"encoding/json"
	"log"
	"resume-backend-service/internal/dto"

	"github.com/gofiber/fiber/v2"
)

type AWSHandler struct{}

func NewAWSHandler() *AWSHandler {
	return &AWSHandler{}
}

func (h *AWSHandler) ProcessResumeResultsHandler(c *fiber.Ctx) error {
	// Log del body raw para debug
	bodyRaw := c.Body()
	log.Printf("📥 Body raw recibido (%d bytes): %s", len(bodyRaw), string(bodyRaw))

	// Log de headers para debug
	log.Printf("📋 Headers recibidos:")
	c.Request().Header.VisitAll(func(key, value []byte) {
		log.Printf("  %s: %s", string(key), string(value))
	})

	var processedData dto.CVProcessedData
	if err := c.BodyParser(&processedData); err != nil {
		log.Printf("❌ Error al parsear el cuerpo de la solicitud: %v", err)

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al parsear el cuerpo de la solicitud.",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// Log del JSON parseado de forma legible
	jsonPretty, _ := json.MarshalIndent(processedData, "", "  ")
	log.Printf("✅ Datos procesados correctamente:\n%s", string(jsonPretty))

	response := dto.AWSProcessResponse{
		Status:  "success",
		Message: "Datos procesados correctamente.",
	}

	return c.Status(fiber.StatusOK).JSON(response)

}
