package handlers

import (
	"resume-backend-service/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ResumeHandler struct {
	resumeService *services.ResumeService
}

func NewResumeHandler(resumeService *services.ResumeService) *ResumeHandler {
	return &ResumeHandler{
		resumeService: resumeService,
	}
}

func (h *ResumeHandler) ProcessResumeHandler(c *fiber.Ctx) error {
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
		language = "esp"
	}

	// Extraer email del token JWT (guardado por el middleware de autenticación)
	userEmail := ""
	if userID := c.Locals("user_id"); userID != nil {
		userEmail = userID.(string)
	}

	response, err := h.resumeService.ProcessResume(
		instructions,
		language,
		userEmail,
		fileHeader,
	)
	if err != nil {
		// Si es un error de Fiber, retornarlo con su código
		if fiberErr, ok := err.(*fiber.Error); ok {
			return c.Status(fiberErr.Code).JSON(fiber.Map{
				"status":  "error",
				"message": fiberErr.Message,
			})
		}
		// Otro tipo de error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Error interno del servidor.",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}
