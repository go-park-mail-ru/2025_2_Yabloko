package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string

	AppPort string

	SecretKey      string
	CSRFSecret     string
	AllowedOrigins string
	CookieSecure   bool
	CookieSameSite string
}

func LoadConfig() *Config {
	return &Config{
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("API_DB_PORT", "5432"),
		DBName:         getEnv("DB_NAME", "postgres"),
		AppPort:        getEnv("AUTH_PORT", "8082"),
		SecretKey:      getEnv("SECRET_KEY", "dev-secret"),
		CSRFSecret:     getEnv("CSRF_SECRET", "dev-csrf-secret"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"),
		CookieSecure:   parseBool(getEnv("COOKIE_SECURE", "false")),
		CookieSameSite: strings.ToLower(getEnv("COOKIE_SAMESITE", "lax")),
	}
}

func (c *Config) DBPath() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func (c *Config) AppPortStr() string   { return c.AppPort }
func (c *Config) SecretKeyStr() string { return c.SecretKey }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	default:
		return false
	}
}
