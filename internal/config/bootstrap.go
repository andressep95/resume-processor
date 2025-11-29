package config

import (
	"log"
	router "resume-backend-service/internal/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Application struct {
	App    *fiber.App
	Config *Config
}

func Bootstrap() *Application {
	// Cargar configuración
	cfg := Load()

	// Crear instancia de Fiber
	app := fiber.New(fiber.Config{
		AppName: "Resume Backend Service",
	})

	// Middlewares globales
	app.Use(logger.New())
	app.Use(recover.New())

	// Registrar rutas
	router.SetupRoutes(app)

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
