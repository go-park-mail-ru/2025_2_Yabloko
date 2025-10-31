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
	AppPort    string

	UploadPath string // локальная папка для аватарок
	BaseURL    string // базовый публичный url для файлов
}

func LoadConfig() *Config {
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "db"
	}

	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &Config{
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBHost:     dbHost,
		DBPort:     os.Getenv("POSTGRES_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("PROFILE_SERVICE_PORT"),
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
