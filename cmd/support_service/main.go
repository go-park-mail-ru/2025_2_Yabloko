package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/support_service/cmd"
	"log/slog"
)

// @title Support Service API + WS
// @version 1.0
// @description Support Service
// @host localhost:8085
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
