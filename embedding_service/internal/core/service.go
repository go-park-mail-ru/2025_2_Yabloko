package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/kelindar/search"
)

type EmbeddingService struct {
	vectorizer *search.Vectorizer
	cache      map[string][]float32
	mu         sync.RWMutex
	logger     *slog.Logger
}

func NewEmbeddingService(modelPath string, logger *slog.Logger) (*EmbeddingService, error) {
	vectorizer, err := search.NewVectorizer(modelPath, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	ctx := vectorizer.Context(128)
	logger.Info("context created, testing embedding...")

	testEmbedding, err := ctx.EmbedText("test")
	if err != nil {
		return nil, fmt.Errorf("failed to get dimensions: %w", err)
	}

	dims := len(testEmbedding)

	logger.Info("embedding model loaded",
		slog.String("model", modelPath),
		slog.Int("dimensions", dims),
	)

	return &EmbeddingService{
		vectorizer: vectorizer,
		cache:      make(map[string][]float32),
		logger:     logger,
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
	embedding, err := e.vectorizer.EmbedText(text)
	if err != nil {
		e.logger.ErrorContext(ctx, "vectorization failed",
			slog.Any("err", err),
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("vectorization failed: %w", err)
	}

	if embedding == nil || len(embedding) == 0 {
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
	return e.vectorizer != nil
}

func (e *EmbeddingService) Close() error {
	if e.vectorizer != nil {
		e.vectorizer.Close()
	}
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
