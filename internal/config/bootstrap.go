package config

import (
	"log"
	"resume-backend-service/internal/middleware"
	router "resume-backend-service/internal/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

type Application struct {
	App    *fiber.App
	Config *Config
}

func Bootstrap() *Application {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No se encontró archivo .env, usando valores por defecto")
	}

	// Cargar configuración
	cfg := Load()

	// Crear instancia de Fiber
	app := fiber.New(fiber.Config{
		AppName: "Resume Backend Service",
	})

	// Middlewares globales
	app.Use(logger.New())
	app.Use(recover.New())

	// Inicializar middleware de autenticación
	authMiddleware := middleware.NewAuthMiddleware(cfg.AuthJWKSURL)

	// Registrar rutas (pasar valores individuales y middleware)
	router.SetupRoutes(app, cfg.PresignedURLServiceEndpoint, authMiddleware)

	return &Application{
		App:    app,
		Config: cfg,
	}
}

func (a *Application) Run() {
	if err := a.App.Listen(":" + a.Config.Port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
