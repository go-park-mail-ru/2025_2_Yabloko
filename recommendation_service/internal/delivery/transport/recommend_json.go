package transport

import "apple_backend/recommendation_service/internal/domain"

type RecommendedItem struct {
	ID      string  `json:"id"`
	StoreID string  `json:"store_id"`
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	CardImg string  `json:"card_img"`
	Score   float64 `json:"score"`
} // @name RecommendedItem

type HomeRecommendationResponse struct {
	Items []*RecommendedItem `json:"items"`
} // @name HomeRecommendationResponse

func toRecommendedItemResponse(item *domain.RecommendedItem) *RecommendedItem {
	return &RecommendedItem{
		ID:      item.ID,
		StoreID: item.StoreID,
		Name:    item.Name,
		Price:   item.Price,
		CardImg: item.CardImg,
		Score:   item.Score,
	}
}

func ToHomeRecommendationResponse(items []*domain.RecommendedItem) *HomeRecommendationResponse {
	respItems := make([]*RecommendedItem, 0, len(items))
	for _, it := range items {
		respItems = append(respItems, toRecommendedItemResponse(it))
	}
	return &HomeRecommendationResponse{
		Items: respItems,
	}
}
