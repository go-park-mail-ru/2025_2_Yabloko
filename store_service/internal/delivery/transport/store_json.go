package transport

import "apple_backend/store_service/internal/domain"

type StoreResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	CityID       string   `json:"city_id"`
	Address      string   `json:"address"`
	CardImg      string   `json:"card_img"`
	Rating       float64  `json:"rating"`
	TagsID       []string `json:"tags_id"`
	CategoriesID []string `json:"categories_id"`
	OpenAt       string   `json:"open_at"`
	ClosedAt     string   `json:"closed_at"`
} // @name StoreResponse

type CityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name CityResponse

type TagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name TagResponse

type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name CategoryResponse

type StoreReview struct {
	UserName  string  `json:"user_name"`
	Rating    float64 `json:"rating"`
	Comment   string  `json:"comment"`
	CreatedAt string  `json:"created_at"`
} // @name StoreReview

// ItemResponse - для search/items endpoint (StoreWithItems)
type ItemResponse struct {
	ID      string   `json:"id"` // store_item.id
	Name    string   `json:"name"`
	Price   float64  `json:"price"`
	TypesID []string `json:"types_id"`
	CardImg string   `json:"card_img"`
} // @name ItemResponse

// ItemAggResponse - для stores/{id}/items endpoint
type ItemAggResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	CardImg     string   `json:"card_img"`
	TypesID     []string `json:"types_id"`
} // @name ItemAggResponse

type ItemTypeResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name ItemTypeResponse

type StoreWithItemsResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	CityID       string          `json:"city_id"`
	Address      string          `json:"address"`
	CardImg      string          `json:"card_img"`
	Rating       float64         `json:"rating"`
	TagsID       []string        `json:"tags_id"`
	CategoriesID []string        `json:"categories_id"`
	OpenAt       string          `json:"open_at"`
	ClosedAt     string          `json:"closed_at"`
	Items        []*ItemResponse `json:"items"`
} // @name StoreWithItemsResponse

// ============ Store conversions ============

func ToStoreResponse(store *domain.StoreAgg) *StoreResponse {
	if store == nil {
		return nil
	}

	return &StoreResponse{
		ID:           store.ID,
		Name:         store.Name,
		Description:  store.Description,
		CityID:       store.CityID,
		Address:      store.Address,
		CardImg:      store.CardImg,
		Rating:       store.Rating,
		TagsID:       store.TagsID,
		CategoriesID: store.CategoriesID,
		OpenAt:       store.OpenAt,
		ClosedAt:     store.ClosedAt,
	}
}

func ToStoreResponses(stores []*domain.StoreAgg) []*StoreResponse {
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

func ToCategoryResponse(category *domain.Category) *CategoryResponse {
	if category == nil {
		return nil
	}

	return &CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
	}
}

func ToCategoryResponses(categories []*domain.Category) []*CategoryResponse {
	responses := make([]*CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, ToCategoryResponse(category))
	}
	return responses
}

// ============ Item conversions ============

// ToItemResponse - для search/items (StoreWithItems)
func ToItemResponse(item *domain.Item) *ItemResponse {
	if item == nil {
		return nil
	}

	return &ItemResponse{
		ID:      item.ID,
		Name:    item.Name,
		Price:   item.Price,
		TypesID: item.TypesID,
		CardImg: item.CardImg,
	}
}

func ToItemResponses(items []*domain.Item) []*ItemResponse {
	responses := make([]*ItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, ToItemResponse(item))
	}
	return responses
}

// ToItemAggResponse - для stores/{id}/items endpoint
func ToItemAggResponse(item *domain.ItemAgg) *ItemAggResponse {
	if item == nil {
		return nil
	}

	return &ItemAggResponse{
		ID:          item.ID,
		Name:        item.Name,
		Price:       item.Price,
		Description: item.Description,
		CardImg:     "/images/items/" + item.CardImg,
		TypesID:     item.TypesID,
	}
}

func ToItemAggResponses(items []*domain.ItemAgg) []*ItemAggResponse {
	responses := make([]*ItemAggResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, ToItemAggResponse(item))
	}
	return responses
}

// ToItemTypeResponse
func ToItemTypeResponse(itemType *domain.ItemType) *ItemTypeResponse {
	if itemType == nil {
		return nil
	}

	return &ItemTypeResponse{
		ID:   itemType.ID,
		Name: itemType.Name,
	}
}

func ToItemTypeResponses(itemTypes []*domain.ItemType) []*ItemTypeResponse {
	responses := make([]*ItemTypeResponse, 0, len(itemTypes))
	for _, itemType := range itemTypes {
		responses = append(responses, ToItemTypeResponse(itemType))
	}
	return responses
}

// ToStoreWithItemsResponse - для search/items endpoint
func ToStoreWithItemsResponse(store *domain.StoreWithItems) *StoreWithItemsResponse {
	if store == nil {
		return nil
	}

	return &StoreWithItemsResponse{
		ID:           store.ID,
		Name:         store.Name,
		Description:  store.Description,
		CityID:       store.CityID,
		Address:      store.Address,
		CardImg:      store.CardImg,
		Rating:       store.Rating,
		TagsID:       store.TagsID,
		CategoriesID: store.CategoriesID,
		OpenAt:       store.OpenAt,
		ClosedAt:     store.ClosedAt,
		Items:        ToItemResponses(store.Items),
	}
}

func ToStoreWithItemsResponses(stores []*domain.StoreWithItems) []*StoreWithItemsResponse {
	responses := make([]*StoreWithItemsResponse, 0, len(stores))
	for _, store := range stores {
		responses = append(responses, ToStoreWithItemsResponse(store))
	}
	return responses
}
