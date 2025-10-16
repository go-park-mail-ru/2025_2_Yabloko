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

type CityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

func ToCityResponse(city *domain.City) *CityResponse {
	if city == nil {
		return nil
	}

	return &CityResponse{
		ID:   city.ID,
		Name: city.Name,
	}
}

// TODO наврное можно дженерики и в pkg
func ToCityResponses(cities []*domain.City) []*CityResponse {
	responses := make([]*CityResponse, 0, len(cities))
	for _, city := range cities {
		responses = append(responses, ToCityResponse(city))
	}
	return responses
}

func ToTagResponse(tag *domain.StoreTag) *TagResponse {
	if tag == nil {
		return nil
	}

	return &TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

func ToTagResponses(tags []*domain.StoreTag) []*TagResponse {
	responses := make([]*TagResponse, 0, len(tags))
	for _, tag := range tags {
		responses = append(responses, ToTagResponse(tag))
	}
	return responses
}
