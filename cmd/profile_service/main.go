package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/cmd"
	"log/slog"
)

// @title Profile Service API
// @version 1.0
// @description Profile Service
// @host localhost:8081
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
