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

func (uc *ItemUsecase) GetItems(ctx context.Context, id string) ([]*domain.ItemAgg, error) {
	items, err := uc.repo.GetItems(ctx, id)
	if err != nil {
		return nil, err
	}

	itemsMap := map[string]*domain.ItemAgg{}
	for _, item := range items {
		// если есть запись по этому товару
		if itemMap, ok := itemsMap[item.ID]; ok {
			itemMap.TypesID = append(itemMap.TypesID, item.TypeID)
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

	itemsList := make([]*domain.ItemAgg, 0, len(itemsMap))
	for _, item := range itemsMap {
		itemsList = append(itemsList, item)
	}

	return itemsList, nil
}
