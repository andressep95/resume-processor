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

	// Configuración de Autenticación
	AuthJWKSURL string

	// Configuración de CORS
	CORSAllowedOrigins string
}

// Load inicializa y retorna la configuración de la aplicación, leyendo
// las variables de entorno o usando valores por defecto.
func Load() *Config {
	// 1. Tamaño por defecto en MB
	const defaultMaxFileSizeMB = 10

	cfg := &Config{
		// 1. Puerto del Servidor (Esencial)
		Port: getEnv("SERVER_PORT", "8080"),

		// 2. Tamaño Máximo de Archivo en MB (se convierte a bytes internamente)
		MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE_MB", defaultMaxFileSizeMB) * 1024 * 1024,

		// 3. Endpoint del Servicio de Presigned URL (ESENCIAL)
		// Requerido para que el handler sepa a dónde llamar para obtener la URL de subida.
		PresignedURLServiceEndpoint: getEnv("PRESIGNED_URL_SERVICE_ENDPOINT", "http://localhost:8081/api/v1/s3/presign"),

		// 4. URL del JWKS para validación de tokens JWT
		AuthJWKSURL: getEnv("AUTH_JWKS_URL", "https://auth.cloudcentinel.com/.well-known/jwks.json"),

		// 5. Orígenes permitidos para CORS (separados por coma)
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
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
