package client

import (
	embeddingv1 "apple_backend/pkg/proto/embedding"
	"apple_backend/store_service/internal/domain"
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCEmbeddingClient struct {
	client embeddingv1.EmbeddingServiceClient
	cache  map[string][]float32
	mu     sync.RWMutex
	logger *slog.Logger
}

func NewGRPCEmbeddingClient(addr string, logger *slog.Logger) (*GRPCEmbeddingClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50*1024*1024)), // 50MB
	)
	if err != nil {
		return nil, err
	}

	return &GRPCEmbeddingClient{
		client: embeddingv1.NewEmbeddingServiceClient(conn),
		cache:  make(map[string][]float32),
		logger: logger,
	}, nil
}

func (c *GRPCEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Проверить кэш
	c.mu.RLock()
	if cached, ok := c.cache[text]; ok {
		c.mu.RUnlock()
		c.logger.DebugContext(ctx, "embedding from cache", slog.Int("dim", len(cached)))
		return cached, nil
	}
	c.mu.RUnlock()

	// С timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := c.client.GetEmbedding(ctx, &embeddingv1.GetEmbeddingRequest{
		Text: text,
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "grpc embedding failed", slog.Any("err", err))
		return nil, err
	}

	embedding := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		embedding[i] = v
	}

	// Кэшировать
	c.mu.Lock()
	c.cache[text] = embedding
	c.mu.Unlock()

	c.logger.DebugContext(ctx, "embedding generated",
		slog.Int("dim", len(embedding)),
		slog.String("model", resp.Model),
	)

	return embedding, nil
}

func (c *GRPCEmbeddingClient) GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := c.client.GetEmbeddingBatch(ctx, &embeddingv1.GetEmbeddingBatchRequest{
		Texts: texts,
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "grpc batch embedding failed", slog.Any("err", err))
		return nil, err
	}

	result := make([][]float32, len(resp.Results))
	for i, r := range resp.Results {
		embedding := make([]float32, len(r.Embedding))
		for j, v := range r.Embedding {
			embedding[j] = v
		}
		result[i] = embedding
	}

	return result, nil
}

type NoOpEmbeddingClient struct {
	logger *slog.Logger
}

func NewNoOpEmbeddingClient(logger *slog.Logger) *NoOpEmbeddingClient {
	return &NoOpEmbeddingClient{logger: logger}
}

func (c *NoOpEmbeddingClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	c.logger.WarnContext(ctx, "embedding service unavailable, using fallback")
	return nil, domain.ErrInternalServer
}

func (c *NoOpEmbeddingClient) GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, domain.ErrInternalServer
}
