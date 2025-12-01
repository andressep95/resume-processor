package config

import (
	"database/sql"
	"log"
	"resume-backend-service/internal/middleware"
	router "resume-backend-service/internal/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

type Application struct {
	App    *fiber.App
	Config *Config
	DB     *sql.DB
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

	// Inicializar base de datos
	db, err := InitDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Error al conectar con base de datos: %v", err)
	}

	// Crear instancia de Fiber
	app := fiber.New(fiber.Config{
		AppName: "Resume Backend Service",
	})

	// Middlewares globales
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	// Inicializar middleware de autenticación
	authMiddleware := middleware.NewAuthMiddleware(cfg.AuthJWKSURL)

	// Registrar rutas (pasar base de datos, configuración y middleware)
	router.SetupRoutes(app, db, cfg.PresignedURLServiceEndpoint, authMiddleware)

	return &Application{
		App:    app,
		Config: cfg,
		DB:     db,
	}
}

func (a *Application) Run() {
	defer a.DB.Close()

	if err := a.App.Listen(":" + a.Config.Port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}
