package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"apple_backend/embedding_service/internal/core"
	embeddingv1 "apple_backend/pkg/proto/embedding"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	embeddingv1.UnimplementedEmbeddingServiceServer
	service *core.EmbeddingService
	logger  *slog.Logger
}

func NewGRPCServer(svc *core.EmbeddingService, logger *slog.Logger) *GRPCServer {
	return &GRPCServer{
		service: svc,
		logger:  logger,
	}
}

func (s *GRPCServer) GetEmbedding(ctx context.Context, req *embeddingv1.GetEmbeddingRequest) (*embeddingv1.GetEmbeddingResponse, error) {
	if req.Text == "" {
		s.logger.WarnContext(ctx, "empty text in GetEmbedding")
		return nil, fmt.Errorf("text is required")
	}

	embedding, err := s.service.GetEmbedding(ctx, req.Text)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetEmbedding failed", slog.Any("err", err))
		return nil, err
	}

	return &embeddingv1.GetEmbeddingResponse{
		Embedding:  embedding,
		Model:      "nomic-embed-text",
		Dimensions: int32(len(embedding)),
	}, nil
}

func (s *GRPCServer) GetEmbeddingBatch(ctx context.Context, req *embeddingv1.GetEmbeddingBatchRequest) (*embeddingv1.GetEmbeddingBatchResponse, error) {
	if len(req.Texts) == 0 {
		s.logger.WarnContext(ctx, "empty texts in GetEmbeddingBatch")
		return nil, fmt.Errorf("texts are required")
	}

	embeddings, err := s.service.GetEmbeddingBatch(ctx, req.Texts)
	if err != nil {
		s.logger.ErrorContext(ctx, "GetEmbeddingBatch failed", slog.Any("err", err))
		return nil, err
	}

	results := make([]*embeddingv1.EmbeddingResult, len(embeddings))
	for i, embedding := range embeddings {
		results[i] = &embeddingv1.EmbeddingResult{
			Text:      req.Texts[i],
			Embedding: embedding,
		}
	}

	return &embeddingv1.GetEmbeddingBatchResponse{
		Results: results,
	}, nil
}

func (s *GRPCServer) Health(ctx context.Context, req *embeddingv1.HealthRequest) (*embeddingv1.HealthResponse, error) {
	healthy := s.service.IsHealthy()
	message := "healthy"
	if !healthy {
		message = "unhealthy"
	}

	return &embeddingv1.HealthResponse{
		Healthy: healthy,
		Message: message,
	}, nil
}

func StartEmbeddingServer(addr string, svc *core.EmbeddingService, logger *slog.Logger) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	server := NewGRPCServer(svc, logger)
	embeddingv1.RegisterEmbeddingServiceServer(grpcServer, server)

	logger.Info("gRPC server starting", slog.String("addr", addr))

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
