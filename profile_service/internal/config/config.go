package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	AppHost    string
	AppPort    string
	JWTSecret  string
	UploadPath string
	BaseURL    string
}

func LoadConfig() *Config {
	// DB
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = os.Getenv("DB_HOST")
	}
	if dbHost == "" {
		dbHost = "db"
	}

	dbPort := os.Getenv("API_DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	// App
	appHost := os.Getenv("PROFILE_SERVICE_HOST")
	if appHost == "" {
		appHost = "localhost"
	}

	appPort := os.Getenv("PROFILE_SERVICE_PORT")
	if appPort == "" {
		appPort = "8081"
	}

	// Build public BaseURL
	baseURL := fmt.Sprintf("http://%s:%s", appHost, appPort)

	// Upload
	uploadPath := os.Getenv("UPLOAD_DIR")
	if uploadPath == "" {
		uploadPath = "/app/avatars"
	}

	return &Config{
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBName:     getEnv("DB_NAME", "postgres"),
		AppHost:    appHost,
		AppPort:    appPort,
		JWTSecret:  os.Getenv("SECRET_KEY"),
		UploadPath: uploadPath,
		BaseURL:    baseURL,
	}
}

func (c *Config) DBPath() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
