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
	appLog := logger.NewLogger("auth_service", slog.LevelInfo)
	accessLog := logger.NewLogger("auth_access", slog.LevelInfo)
	cmd.Run(appLog, accessLog)
}
