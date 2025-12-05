package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

type EmbeddingService struct {
	serverURL  string
	httpClient *http.Client
	cache      map[string][]float32
	mu         sync.RWMutex
	logger     *slog.Logger
}

func NewEmbeddingService(serverURL string, logger *slog.Logger) (*EmbeddingService, error) {
	logger.Info("connecting to llama.cpp server", slog.String("url", serverURL))

	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("server not available: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server unhealthy: status %d", resp.StatusCode)
	}

	logger.Info("connected to llama.cpp server successfully")

	return &EmbeddingService{
		serverURL:  serverURL,
		httpClient: &http.Client{},
		cache:      make(map[string][]float32),
		logger:     logger,
	}, nil
}

type embeddingRequest struct {
	Content string `json:"content"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
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

	reqBody := embeddingRequest{Content: text}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := e.httpClient.Post(
		e.serverURL+"/embedding",
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		e.logger.ErrorContext(ctx, "embedding request failed",
			slog.Any("err", err),
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		e.logger.ErrorContext(ctx, "failed to parse embedding response",
			slog.Any("err", err),
		)
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if len(embResp.Embedding) == 0 {
		e.logger.ErrorContext(ctx, "embedding returned empty result",
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("empty embedding")
	}

	// Cache result
	e.mu.Lock()
	e.cache[text] = embResp.Embedding
	e.mu.Unlock()

	e.logger.DebugContext(ctx, "embedding generated",
		slog.Int("dim", len(embResp.Embedding)),
	)

	return embResp.Embedding, nil
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
	resp, err := e.httpClient.Get(e.serverURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (e *EmbeddingService) Close() error {
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
