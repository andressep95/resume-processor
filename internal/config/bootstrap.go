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
	// Cargar variables de entorno desde .env (solo en desarrollo)
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  Usando variables de entorno del sistema (producción)")
	}

	// Cargar configuración
	cfg := Load()
	log.Printf("✅ Configuración cargada: Port=%s, MaxFileSize=%dMB, AuthEnabled=%v",
		cfg.Port, cfg.MaxFileSize/(1024*1024), cfg.AuthJWKSURL != "")

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
