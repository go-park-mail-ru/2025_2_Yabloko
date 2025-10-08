package transport

import "apple_backend/store_service/internal/domain"

type StoreResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CityID      string  `json:"city_id"`
	Address     string  `json:"address"`
	CardImg     string  `json:"card_img"`
	Rating      float64 `json:"rating"`
	OpenAt      string  `json:"open_at"`
	ClosedAt    string  `json:"closed_at"`
}

func ToStoreResponse(store *domain.Store) *StoreResponse {
	if store == nil {
		return nil
	}

	return &StoreResponse{
		ID:          store.ID,
		Name:        store.Name,
		Description: store.Description,
		CityID:      store.CityID,
		Address:     store.Address,
		CardImg:     store.CardImg,
		Rating:      store.Rating,
		OpenAt:      store.OpenAt,
		ClosedAt:    store.ClosedAt,
	}
}

func ToStoreResponses(stores []*domain.Store) []*StoreResponse {
	responses := make([]*StoreResponse, 0, len(stores))
	for _, store := range stores {
		responses = append(responses, ToStoreResponse(store))
	}
	return responses
}
