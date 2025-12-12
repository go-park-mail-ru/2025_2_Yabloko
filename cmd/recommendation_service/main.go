package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/recommendation_service/cmd"
	"log/slog"
)

// @title Recommendation Service API
// @version 1.0
// @description Recommendation Service
// @host localhost:8085
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
