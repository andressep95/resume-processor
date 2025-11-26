package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server  ServerConfig
	App     AppConfig
	Storage StorageConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
}

type AppConfig struct {
	Name        string
	Environment string
}

type StorageConfig struct {
	UploadPath   string
	MaxFileSize  int64
	AllowedTypes []string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 10),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
		},
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Resume Backend Service"),
			Environment: getEnv("APP_ENV", "development"),
		},
		Storage: StorageConfig{
			UploadPath:   getEnv("UPLOAD_PATH", "./uploads"),
			MaxFileSize:  getEnvAsInt64("MAX_FILE_SIZE", 10*1024*1024), // 10MB
			AllowedTypes: []string{".txt", ".doc", ".docx", ".pdf"},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
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
