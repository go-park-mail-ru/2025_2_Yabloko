package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DBUser     string `validate:"required"`
	DBPassword string `validate:"required"`
	DBHost     string `validate:"required"`
	DBPort     string `validate:"required"`
	DBName     string `validate:"required"`
	AppPort    string `validate:"required"`
}

func MustConfig() *Config {
	conf := &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("API_DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("RECOMMENDATION_SERVICE_PORT"),
	}

	if err := validator.New().Struct(conf); err != nil {
		panic(fmt.Sprintf("Некорректно заполнен файл .env %v", err))
	}

	return conf
}

func (c *Config) DBPath() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
