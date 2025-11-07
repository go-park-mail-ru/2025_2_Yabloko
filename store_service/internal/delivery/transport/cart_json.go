package transport

import "apple_backend/store_service/internal/domain"

type CartItem struct {
	// ID из таблицы store_item
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	CardImg  string  `json:"card_img"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
} // @name CartItem

type Cart struct {
	Items []*CartItem `json:"items"`
} // @name Cart

type ItemUpdate struct {
	// id - store_item_id
	ID       string `json:"id" validate:"required, uuid"`
	Quantity int    `json:"quantity" validate:"required"`
} // @name ItemUpdate

type CartUpdate struct {
	Items []*ItemUpdate `json:"items" validate:"required"`
} // @name CartUpdate

func toCartItemResponse(item *domain.CartItem) *CartItem {
	return &CartItem{
		ID:       item.ID,
		Name:     item.Name,
		CardImg:  "/images/stores/" + item.CardImg,
		Price:    item.Price,
		Quantity: item.Quantity,
	}
}

func ToCartResponse(cart *domain.Cart) *Cart {
	items := make([]*CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, toCartItemResponse(item))
	}
	respCart := &Cart{
		Items: items,
	}

	return respCart
}

func FromCartUpdate(cartRequest *CartUpdate) *domain.CartUpdate {
	cartItems := make([]*domain.ItemUpdate, 0, len(cartRequest.Items))
	for _, item := range cartRequest.Items {
		cartItems = append(cartItems, &domain.ItemUpdate{ID: item.ID, Quantity: item.Quantity})
	}
	cartUpdate := &domain.CartUpdate{Items: cartItems}

	return cartUpdate
}
