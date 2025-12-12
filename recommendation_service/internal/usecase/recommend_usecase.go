package usecase

import (
	"apple_backend/recommendation_service/internal/domain"
	"context"
	"errors"
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
	if filter == nil {
		return nil, domain.ErrRequestParams
	}
	if filter.Limit <= 0 || filter.Limit > 20 {
		return nil, domain.ErrRequestParams
	}
	if filter.UserID == "" {
		return nil, domain.ErrRequestParams
	}

	items, err := uc.repo.GetHomeRecommendations(ctx, filter)
	if err != nil {
		if errors.Is(err, domain.ErrInternalServer) {
			return nil, domain.ErrInternalServer
		}
		return nil, domain.ErrInternalServer
	}

	for _, it := range items {
		it.Score = math.Round(it.Score*1000) / 1000
	}

	return items, nil
}
