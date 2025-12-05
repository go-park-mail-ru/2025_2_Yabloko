package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	GRPCPort  string `validate:"required"`
	ModelPath string `validate:"required"`
	LogLevel  string `validate:"required,oneof=debug info warn error"`
}

func MustConfig() *Config {
	conf := &Config{
		GRPCPort:  os.Getenv("GRPC_PORT"),
		ModelPath: os.Getenv("MODEL_PATH"),
		LogLevel:  os.Getenv("LOG_LEVEL"),
	}

	if conf.GRPCPort == "" {
		conf.GRPCPort = "50051"
	}
	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}

	if err := validator.New().Struct(conf); err != nil {
		panic(fmt.Sprintf("Invalid config: %v", err))
	}

	return conf
}
