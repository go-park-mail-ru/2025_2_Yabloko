package main

import (
	"apple_backend/auth_service/cmd"
	"apple_backend/pkg/logger"
	"log/slog"
)

// @title Auth Service API
// @version 1.0
// @description Authentication Service
// @host localhost:8082
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)

	cmd.Run()
}
