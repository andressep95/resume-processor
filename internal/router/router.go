package router

import (
	"database/sql"
	"resume-backend-service/internal/handlers"
	"resume-backend-service/internal/middleware"
	"resume-backend-service/internal/repository"
	"resume-backend-service/internal/services"
	"resume-backend-service/pkg/client"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, db *sql.DB, presignedURLEndpoint string, authMiddleware *middleware.AuthMiddleware) {
	// API v1
	api := app.Group("/api/v1")

	// Health routes (sin autenticación)
	health := api.Group("/health")
	healthHandler := handlers.NewHealthHandler()
	health.Get("/", healthHandler.HandleHealthCheck)

	// Inicializar repositorios
	resumeRequestRepo := repository.NewResumeRequestRepository(db)
	processedResumeRepo := repository.NewProcessedResumeRepository(db)

	// Inicializar clientes
	presignedURLClient := client.NewPresignedURLClient(presignedURLEndpoint)

	// Inicializar servicios
	resumeService := services.NewResumeService(presignedURLClient, resumeRequestRepo)

	// Inicializar handlers con dependencias
	resumeHandler := handlers.NewResumeHandler(resumeService)
	awsHandler := handlers.NewAWSHandler(resumeRequestRepo, processedResumeRepo)

	// CV Processor routes
	resume := api.Group("/resume")

	// Endpoint protegido (requiere autenticación de usuario)
	resume.Post("/", authMiddleware.ValidateJWT(), resumeHandler.ProcessResumeHandler)

	// Endpoint público (callback de AWS Lambda)
	resume.Post("/results", awsHandler.ProcessResumeResultsHandler)

}
