package main

import (
	"apple_backend/order_service/cmd"
	"apple_backend/pkg/logger"
	"log/slog"
)

// @title Order Service API
// @version 1.0
// @description Order Service
// @host localhost:8084
// @BasePath /api/v0
func main() {
	_ = logger.NewLogger("", slog.LevelInfo)
	cmd.Run()
}
