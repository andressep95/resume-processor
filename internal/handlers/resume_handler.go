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

	// Extraer user_id del token JWT (guardado por el middleware de autenticación)
	// El middleware guarda el subject (UUID del usuario) en user_subject
	userID := ""
	if subject := c.Locals("user_subject"); subject != nil {
		userID = subject.(string)
	}

	// Si no hay user_id en el token, rechazar la solicitud
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "No se pudo identificar al usuario.",
		})
	}

	response, err := h.resumeService.ProcessResume(
		userID,
		instructions,
		language,
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
