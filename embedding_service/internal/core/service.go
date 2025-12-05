package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	llama "github.com/go-skynet/go-llama.cpp"
)

type EmbeddingService struct {
	model  *llama.LLama
	cache  map[string][]float32
	mu     sync.RWMutex
	logger *slog.Logger
}

func NewEmbeddingService(modelPath string, logger *slog.Logger) (*EmbeddingService, error) {
	logger.Info("loading model...", slog.String("model", modelPath))

	// Загружаем модель с поддержкой embeddings
	model, err := llama.New(
		modelPath,
		llama.SetContext(512),  // Размер контекста
		llama.SetThreads(1),    // 1 CPU
		llama.EnableEmbeddings, // Включаем embeddings mode
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	logger.Info("model loaded successfully")

	return &EmbeddingService{
		model:  model,
		cache:  make(map[string][]float32),
		logger: logger,
	}, nil
}

func (e *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	// Check cache
	e.mu.RLock()
	if cached, ok := e.cache[text]; ok {
		e.mu.RUnlock()
		e.logger.DebugContext(ctx, "embedding from cache",
			slog.Int("dim", len(cached)),
		)
		return cached, nil
	}
	e.mu.RUnlock()

	// Generate embedding
	embedding, err := e.model.Embeddings(text)
	if err != nil {
		e.logger.ErrorContext(ctx, "vectorization failed",
			slog.Any("err", err),
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("vectorization failed: %w", err)
	}

	if len(embedding) == 0 {
		e.logger.ErrorContext(ctx, "vectorization returned empty result",
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("vectorization failed: empty result")
	}

	// Cache result
	e.mu.Lock()
	e.cache[text] = embedding
	e.mu.Unlock()

	e.logger.DebugContext(ctx, "embedding generated",
		slog.Int("dim", len(embedding)),
	)

	return embedding, nil
}

func (e *EmbeddingService) GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty texts")
	}

	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := e.GetEmbedding(ctx, text)
		if err != nil {
			e.logger.ErrorContext(ctx, "batch embedding failed",
				slog.Any("err", err),
				slog.Int("index", i),
			)
			return nil, err
		}
		results[i] = embedding
	}

	e.logger.InfoContext(ctx, "batch embedding completed",
		slog.Int("count", len(texts)),
	)

	return results, nil
}

func (e *EmbeddingService) IsHealthy() bool {
	return e.model != nil
}

func (e *EmbeddingService) Close() error {
	if e.model != nil {
		e.model.Free()
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
