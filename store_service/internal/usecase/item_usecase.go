package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type ItemRepository interface {
	GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error)
	GetItems(ctx context.Context, storeID string) ([]*domain.ItemAgg, error)
}

type ItemUsecase struct {
	repo ItemRepository
}

func NewItemUsecase(repo ItemRepository) *ItemUsecase {
	return &ItemUsecase{repo: repo}
}

func (uc *ItemUsecase) GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error) {
	return uc.repo.GetItemTypes(ctx, storeID)
}

func (uc *ItemUsecase) GetItems(ctx context.Context, storeID string) ([]*domain.ItemAgg, error) {
	return uc.repo.GetItems(ctx, storeID)
}
