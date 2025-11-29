package handlers

import (
	"log"
	"resume-backend-service/internal/dto"

	"github.com/gofiber/fiber/v2"
)

type AWSHandler struct{}

func NewAWSHandler() *AWSHandler {
	return &AWSHandler{}
}

func (h *AWSHandler) ProcessCVHandler(c *fiber.Ctx) error {

	var processedData dto.CVProcessedData
	if err := c.BodyParser(&processedData); err != nil {
		log.Printf("Error al parsear el cuerpo de la solicitud: %v", err)

		response := dto.AWSProcessResponse{
			Status:  "error",
			Message: "Error al parsear el cuerpo de la solicitud.",
		}

		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	log.Printf("Datos procesados correctamente: %+v", processedData)

	response := dto.AWSProcessResponse{
		Status:  "success",
		Message: "Datos procesados correctamente.",
	}

	return c.Status(fiber.StatusOK).JSON(response)

}
