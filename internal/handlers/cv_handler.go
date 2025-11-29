package handlers

import (
	"log"
	"path/filepath"
	"resume-backend-service/internal/dto"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type CVHandler struct {
}

func NewCVHandler() *CVHandler {
	return &CVHandler{}
}

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".doc":  true,
	".docx": true,
}

func (h *CVHandler) ProcessCVHandler(c *fiber.Ctx) error {
	instructions := c.FormValue("instructions")
	language := c.FormValue("language")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Campo 'file' requerido.",
		})
	}

	if language == "" {
		language = "es"
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExtensions[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Formato de archivo no permitido. Permite: .pdf, .txt, .doc, .docx.",
		})
	}

	log.Printf("Solicitud encolada: %s, Instrucciones: %s, Idioma: %s", fileHeader.Filename, instructions, language)

	response := dto.CVProcessorResponseDTO{
		Status:  "accepted",
		Message: "Solicitud encolada para procesamiento.",
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}
