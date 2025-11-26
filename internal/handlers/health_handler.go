package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// NewHealthHandler crea una nueva instancia del handler de health
func NewHealthHandler() *HealthResponse {
	return &HealthResponse{}
}

// HandleHealthCheck maneja las peticiones de health check usando Fiber
func (h *HealthResponse) HandleHealthCheck(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:  "healthy",
		Service: "resume-backend-service",
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
