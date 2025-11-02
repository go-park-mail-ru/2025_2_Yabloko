package config

import "os"

type Config struct {
	Port      string
	JWTSecret string
	DBURL     string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8081"),
		JWTSecret: getEnv("JWT_SECRET", "dev-secret"),
		DBURL:     getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/appdb?sslmode=disable"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
