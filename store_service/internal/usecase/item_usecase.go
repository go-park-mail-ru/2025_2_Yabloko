package usecase

import (
	"apple_backend/store_service/internal/domain"
	"context"
)

type ItemRepository interface {
	GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error)
	GetItems(ctx context.Context, itemTypeID string) ([]*domain.Item, error)
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

func (uc *ItemUsecase) GetItems(ctx context.Context, itemTypeID string) ([]*domain.ItemAgg, error) {
	items, err := uc.repo.GetItems(ctx, itemTypeID)
	if err != nil {
		return nil, err
	}

	// Группировка по ID товара (если один товар в нескольких типах)
	itemsMap := make(map[string]*domain.ItemAgg)
	for _, item := range items {
		if agg, exists := itemsMap[item.ID]; exists {
			agg.TypesID = append(agg.TypesID, item.TypeID)
		} else {
			itemsMap[item.ID] = &domain.ItemAgg{
				ID:          item.ID,
				Name:        item.Name,
				Price:       item.Price,
				Description: item.Description,
				CardImg:     item.CardImg,
				TypesID:     []string{item.TypeID},
			}
		}
	}

	result := make([]*domain.ItemAgg, 0, len(itemsMap))
	for _, agg := range itemsMap {
		result = append(result, agg)
	}
	return result, nil
}
