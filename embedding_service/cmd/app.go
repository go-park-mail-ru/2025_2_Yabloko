package cmd

import (
	"fmt"
	"os"

	"apple_backend/embedding_service/internal/config"
	"apple_backend/embedding_service/internal/core"
	"apple_backend/embedding_service/internal/delivery/grpc"
	"apple_backend/pkg/logger"
)

func Run() {
	conf := config.MustConfig()
	log := logger.Global()

	log.Info("embedding service starting",
		"llama_server", conf.LlamaServerURL,
		"port", conf.GRPCPort,
	)

	embeddingService, err := core.NewEmbeddingService(conf.LlamaServerURL, log)
	if err != nil {
		log.Error("failed to connect to llama server", "err", err)
		os.Exit(1)
	}
	defer embeddingService.Close()

	addr := fmt.Sprintf("0.0.0.0:%s", conf.GRPCPort)
	if err := grpc.StartEmbeddingServer(addr, embeddingService, log); err != nil {
		log.Error("server failed", "err", err)
		os.Exit(1)
	}
}
