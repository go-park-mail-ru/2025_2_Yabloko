package transport

import (
	"apple_backend/order_service/internal/domain"
	"time"
)

type OrderItemInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	CardImg  string  `json:"card_img"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
} // @name OrderItemInfo

type StoreInfo struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	CardImg string           `json:"card_img"`
	Items   []*OrderItemInfo `json:"items"`
} // @name StoreInfo

type OrderInfo struct {
	ID        string       `json:"id"`
	Stores    []*StoreInfo `json:"stores"`
	Status    string       `json:"status"`
	Total     float64      `json:"total"`
	IsFast    bool         `json:"is_fast"`
	Comment   string       `json:"comment"`
	CreatedAt time.Time    `json:"created_at"`
} // @name OrderInfo

type Order struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	StoreID   string    `json:"store_id"`
	StoreName string    `json:"store_name"`
	IsFast    bool      `json:"is_fast"`
	Comment   string    `json:"comment"`
} // @name Order

type Orders struct {
	Orders []*Order `json:"orders"`
} // @name Orders

type OrderStatus struct {
	Status string `json:"status" validate:"required"`
} // @name OrderStatus

type OrderCreateRequest struct {
	IsFast  bool   `json:"is_fast"`
	Comment string `json:"comment"`
	Promo   string `json:"promo"`
}

func toOrderResponse(order *domain.Order) *Order {
	return &Order{
		ID:        order.ID,
		Status:    order.Status,
		Total:     order.Total,
		CreatedAt: order.CreatedAt,
		StoreID:   order.StoreID,
		StoreName: order.StoreName,
		IsFast:    order.IsFast,
		Comment:   order.Comment,
	}
}

func ToOrdersResponse(orders []*domain.Order) []*Order {
	ordersList := make([]*Order, 0, len(orders))
	for _, order := range orders {
		ordersList = append(ordersList, toOrderResponse(order))
	}
	return ordersList
}

func toOrderItemResponse(item *domain.OrderItemInfo) *OrderItemInfo {
	return &OrderItemInfo{
		ID:       item.ID,
		Name:     item.Name,
		CardImg:  item.CardImg,
		Price:    item.Price,
		Quantity: item.Quantity,
	}
}

func ToOrderInfoResponse(orderInfo *domain.OrderInfo) *OrderInfo {
	stores := make([]*StoreInfo, 0, len(orderInfo.Stores))
	for _, store := range orderInfo.Stores {
		items := make([]*OrderItemInfo, 0, len(store.Items))
		for _, item := range store.Items {
			items = append(items, toOrderItemResponse(item))
		}
		stores = append(stores, &StoreInfo{
			ID:      store.ID,
			Name:    store.Name,
			CardImg: store.CardImg,
			Items:   items,
		})
	}

	return &OrderInfo{
		ID:        orderInfo.ID,
		Stores:    stores,
		Status:    orderInfo.Status,
		Total:     orderInfo.Total,
		IsFast:    orderInfo.IsFast,
		Comment:   orderInfo.Comment,
		CreatedAt: orderInfo.CreatedAt,
	}
}
