package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/cmd"
	"log/slog"
)

func main() {
	appLog := logger.NewLogger("./logs/app.log", slog.LevelDebug)
	accessLog := logger.NewLogger("./logs/access.log", slog.LevelDebug)

	cmd.Run(appLog, accessLog)
}
