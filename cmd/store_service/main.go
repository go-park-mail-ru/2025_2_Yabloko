package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/cmd"
	"log/slog"
)

func main() {
	accessLog := logger.NewLogger("../../logs/access.log", slog.LevelDebug)
	appLog := logger.NewLogger("../../logs/app.log", slog.LevelDebug)
	cmd.Run(appLog, accessLog)
}
