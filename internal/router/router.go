package router

import (
	"resume-backend-service/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// API v1
	api := app.Group("/api/v1")

	// Health routes
	health := api.Group("/health")
	healthHandler := handlers.NewHealthHandler()
	health.Get("/", healthHandler.HandleHealthCheck)

	// CV Processor routes
	resume := api.Group("/resume")
	cvHandler := handlers.NewCVHandler()
	resume.Post("/", cvHandler.ProcessCVHandler)
}
