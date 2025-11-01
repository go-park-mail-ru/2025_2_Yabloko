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

	AppPort string `validate:"required"`

	JWTSecret string `validate:"required"`

	ImageDir string `validate:"required"`
}

func MustConfig() *Config {
	conf := &Config{
		DBUser:     os.Getenv("POSTGRES_USER"),
		DBPassword: os.Getenv("POSTGRES_PASSWORD"),
		DBHost:     os.Getenv("POSTGRES_HOST"),
		DBPort:     os.Getenv("POSTGRES_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("STORE_PORT"),
		JWTSecret:  os.Getenv("SECRET_KEY"),
		ImageDir:   os.Getenv("IMAGE_PATH"),
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
