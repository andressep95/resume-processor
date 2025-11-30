package router

import (
	"resume-backend-service/internal/handlers"
	"resume-backend-service/internal/middleware"
	"resume-backend-service/internal/services"
	"resume-backend-service/pkg/client"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, presignedURLEndpoint string, authMiddleware *middleware.AuthMiddleware) {
	// API v1
	api := app.Group("/api/v1")

	// Health routes (sin autenticación)
	health := api.Group("/health")
	healthHandler := handlers.NewHealthHandler()
	health.Get("/", healthHandler.HandleHealthCheck)

	// CV Processor routes (con autenticación)
	resume := api.Group("/resume")
	resume.Use(authMiddleware.ValidateJWT()) // Aplicar middleware de autenticación

	// Inicializar clientes
	presignedURLClient := client.NewPresignedURLClient(presignedURLEndpoint)

	// Inicializar servicios
	resumeService := services.NewResumeService(presignedURLClient)

	// Inicializar handlers con dependencias
	resumeHandler := handlers.NewResumeHandler(resumeService)
	awsHandler := handlers.NewAWSHandler()

	resume.Post("/", resumeHandler.ProcessResumeHandler)
	resume.Post("/results", awsHandler.ProcessResumeResultsHandler)

}
