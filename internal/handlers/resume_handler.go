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

	// Extraer user_id y email del token JWT (guardado por el middleware de autenticación)
	userID := ""
	userEmail := ""
	if id := c.Locals("user_id"); id != nil {
		userID = id.(string)
	}
	if email := c.Locals("user_email"); email != nil {
		userEmail = email.(string)
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
