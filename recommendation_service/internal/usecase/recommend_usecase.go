package usecase

import (
	"apple_backend/pkg/logger"
	"apple_backend/recommendation_service/internal/domain"
	"context"
	"log/slog"
	"math"
)

type RecommendationRepository interface {
	GetHomeRecommendations(ctx context.Context, filter *domain.HomeRecommendFilter) ([]*domain.RecommendedItem, error)
}

type RecommendationUsecase struct {
	repo RecommendationRepository
}

func NewRecommendationUsecase(repo RecommendationRepository) *RecommendationUsecase {
	return &RecommendationUsecase{
		repo: repo,
	}
}

func (uc *RecommendationUsecase) GetHomeRecommendations(ctx context.Context, filter *domain.HomeRecommendFilter) ([]*domain.RecommendedItem, error) {
	if filter == nil || filter.Limit <= 0 || filter.Limit > 20 || filter.UserID == "" {
		logger.FromContext(ctx).ErrorContext(ctx, "uc GetHomeRecommendations invalid filter",
			slog.Any("filter", filter))
		return nil, domain.ErrRequestParams
	}

	items, err := uc.repo.GetHomeRecommendations(ctx, filter)
	if err != nil {
		logger.FromContext(ctx).ErrorContext(ctx, "uc GetHomeRecommendations repo failed",
			slog.Any("err", err))
		return nil, err
	}

	for _, it := range items {
		it.Score = math.Round(it.Score*1000) / 1000
	}

	return items, nil
}
