package main

import (
	"apple_backend/auth_service/cmd"
	"apple_backend/pkg/logger"
	"log/slog"
)

func main() {
	appLog := logger.NewLogger("auth_service", slog.LevelInfo)
	accessLog := logger.NewLogger("auth_access", slog.LevelInfo)
	cmd.Run(appLog, accessLog)
}
