package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret string
}

var cfg *Config

func GetConfig() *Config {
	if cfg == nil {
		err := godotenv.Load()

		if err != nil {
			panic(err.Error())
		}
		fmt.Println(os.Getenv("JWT_SECRET"))
		cfg = &Config{JWTSecret: os.Getenv("JWT_SECRET")}
	}

	return cfg

}
