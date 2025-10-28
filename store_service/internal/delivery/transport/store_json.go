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
} // @name StoreResponse

type CityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name CityResponse

type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name TagResponse

type StoreReview struct {
	UserName  string  `json:"user_name"`
	Rating    float64 `json:"rating"`
	Comment   string  `json:"comment"`
	CreatedAt string  `json:"created_at"`
} // @name StoreReview

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

func ToStoreReview(review *domain.StoreReview) *StoreReview {
	if review == nil {
		return nil
	}

	return &StoreReview{
		UserName:  review.UserName,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
	}
}

func ToStoreReviews(reviews []*domain.StoreReview) []*StoreReview {
	responses := make([]*StoreReview, 0, len(reviews))
	for _, review := range reviews {
		responses = append(responses, ToStoreReview(review))
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
