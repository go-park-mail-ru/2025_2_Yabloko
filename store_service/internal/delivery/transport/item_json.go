package transport

import "apple_backend/store_service/internal/domain"

type Item struct {
	// ID из таблицы store_item
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	CardImg     string   `json:"card_img"`
	TypesID     []string `json:"types_id"`
}

type ItemType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func toItemTypeResponse(itemType *domain.ItemType) *ItemType {
	return &ItemType{
		ID:   itemType.ID,
		Name: itemType.Name,
	}
}

func ToItemTypesResponse(itemTypes []*domain.ItemType) []*ItemType {
	responses := make([]*ItemType, 0, len(itemTypes))
	for _, itemType := range itemTypes {
		responses = append(responses, toItemTypeResponse(itemType))
	}
	return responses
}

func toItemResponse(item *domain.ItemAgg) *Item {
	return &Item{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		CardImg:     item.CardImg,
		TypesID:     item.TypesID,
	}
}

func ToItemsResponse(items []*domain.ItemAgg) []*Item {
	responses := make([]*Item, 0, len(items))
	for _, item := range items {
		responses = append(responses, toItemResponse(item))
	}
	return responses
}
