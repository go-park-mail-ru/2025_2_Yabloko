package main

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/cmd"
	"log/slog"
)

// @title Store Service API
// @version 1.0
// @description Store Service
// @host localhost:8080
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
