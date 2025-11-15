package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DBUser       string `validate:"required"`
	DBPassword   string `validate:"required"`
	DBHost       string `validate:"required"`
	DBPort       string `validate:"required"`
	DBName       string `validate:"required"`
	AppPort      string `validate:"required"`
	JWTSecret    string `validate:"required"`
	AdminUserIDs []string
}

func MustConfig() *Config {
	conf := &Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		AppPort:    os.Getenv("SUPPORT_SERVICE_PORT"),
		JWTSecret:  os.Getenv("SECRET_KEY"),
	}

	if admins := os.Getenv("ADMIN_USER_IDS"); admins != "" {
		conf.AdminUserIDs = strings.Split(admins, ",")
	}

	if err := validator.New().Struct(conf); err != nil {
		panic(fmt.Sprintf("Некорректно заполнен .env: %v", err))
	}

	return conf
}

func (c *Config) DBPath() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}
