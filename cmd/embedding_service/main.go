package main

import (
	"apple_backend/embedding_service/cmd"
	"apple_backend/pkg/logger"
	"log/slog"
)

func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
