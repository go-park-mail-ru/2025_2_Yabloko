package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type ItemRepository interface {
	GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error)
	GetItems(ctx context.Context, filter *domain.ItemFilter) ([]*domain.ItemAgg, error)
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

func (uc *ItemUsecase) GetItems(ctx context.Context, storeID string, itemTypes []string, sorted string, desc bool) ([]*domain.ItemAgg, error) {
	// валидация сортировки
	sortable := map[string]bool{
		"name":  true,
		"price": true,
	}
	if sorted != "" && !sortable[sorted] {
		return nil, domain.ErrRequestParams
	}

	filter := &domain.ItemFilter{
		StoreID:   storeID,
		ItemTypes: itemTypes,
		Sorted:    sorted,
		Desc:      desc,
	}

	items, err := uc.repo.GetItems(ctx, filter)
	if err != nil {
		if err == domain.ErrRowsNotFound {
			return []*domain.ItemAgg{}, nil
		}
		return nil, err
	}
	return items, nil
}
