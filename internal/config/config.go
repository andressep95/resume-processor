package config

import (
	"os"
	"strconv"
)

// Config contiene todos los parámetros esenciales para la aplicación.
type Config struct {
	// Configuración del Servidor
	Port string

	// Configuración de Almacenamiento/Archivos
	MaxFileSize int64

	// Configuración de Servicios Externos
	PresignedURLServiceEndpoint string
}

// Load inicializa y retorna la configuración de la aplicación, leyendo
// las variables de entorno o usando valores por defecto.
func Load() *Config {
	// 1. Definimos el límite de 10MB para la carga de archivos.
	const defaultMaxFileSize = 10485760

	cfg := &Config{
		// 1. Puerto del Servidor (Esencial)
		Port: getEnv("SERVER_PORT", "8080"),

		// 2. Tamaño Máximo de Archivo (Usado en el middleware de Fiber)
		MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", defaultMaxFileSize),

		// 3. Endpoint del Servicio de Presigned URL (ESENCIAL)
		// Requerido para que el handler sepa a dónde llamar para obtener la URL de subida.
		PresignedURLServiceEndpoint: getEnv("PRESIGNED_URL_SERVICE_ENDPOINT", "http://localhost:8081/api/v1/s3/presign"),
	}

	return cfg
}

// --- Funciones de Utilidad ---

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}
