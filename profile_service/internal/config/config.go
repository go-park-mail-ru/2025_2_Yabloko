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
}

func LoadConfig() *Config {
	return &Config{
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBHost:     "db",
		DBPort:     os.Getenv("POSTGRES_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("PROFILE_SERVICE_PORT"),
	}
}

func (c *Config) DBPath() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
