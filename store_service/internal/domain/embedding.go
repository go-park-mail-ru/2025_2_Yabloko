package domain

import "context"

type EmbeddingClient interface {
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
	GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error)
}

type SearchResult struct {
	StoreID       string
	BM25Score     float64
	SemanticScore float64
	CombinedScore float64
}
