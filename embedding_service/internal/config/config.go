package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	GRPCPort  string `validate:"required"`
	ModelPath string `validate:"required"`
	VocabPath string `validate:"required"`
	LogLevel  string `validate:"required,oneof=debug info warn error"`
}

func MustConfig() *Config {
	conf := &Config{
		GRPCPort:  os.Getenv("GRPC_PORT"),
		ModelPath: os.Getenv("MODEL_PATH"),
		VocabPath: os.Getenv("VOCAB_PATH"),
		LogLevel:  os.Getenv("LOG_LEVEL"),
	}

	if conf.GRPCPort == "" {
		conf.GRPCPort = "50051"
	}
	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}
	if conf.ModelPath == "" {
		conf.ModelPath = "models/model.onnx"
	}
	if conf.VocabPath == "" {
		conf.VocabPath = "models/vocab.txt"
	}

	if err := validator.New().Struct(conf); err != nil {
		panic(fmt.Sprintf("Invalid config: %v", err))
	}

	return conf
}
