package config

import (
	"log"
	router "resume-backend-service/internal/routes"

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
		AppName: cfg.App.Name,
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
	addr := a.Config.Server.Host + ":" + a.Config.Server.Port
	log.Printf("🚀 Servidor iniciado en http://%s", addr)
	log.Printf("📝 Ambiente: %s", a.Config.App.Environment)

	if err := a.App.Listen(":" + a.Config.Server.Port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
