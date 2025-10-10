package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type ItemRepository interface {
	// GetItemTypes метод для получения всех типов товаров в конкертном магазине
	GetItemTypes(ctx context.Context, id string) ([]*domain.ItemType, error)
	// GetItems метод получения всех товаров магазина
	GetItems(ctx context.Context, id string) ([]*domain.Item, error)
}

type ItemUsecase struct {
	repo ItemRepository
}

func NewItemUsecase(repo ItemRepository) *ItemUsecase {
	return &ItemUsecase{repo: repo}
}

func (uc *ItemUsecase) GetItemTypes(ctx context.Context, id string) ([]*domain.ItemType, error) {
	return uc.repo.GetItemTypes(ctx, id)
}

func (uc *ItemUsecase) GetItems(ctx context.Context, id string) ([]*domain.Item, error) {
	return uc.repo.GetItems(ctx, id)
}
