package transport

import "apple_backend/store_service/internal/domain"

type OrderItemInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	CardImg  string  `json:"card_img"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
} // @name OrderItemInfo

type OrderInfo struct {
	ID        string           `json:"id"`
	Items     []*OrderItemInfo `json:"items"`
	Status    string           `json:"status"`
	Total     float64          `json:"total"`
	CreatedAt string           `json:"created_at"`
} // @name OrderInfo

type Order struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"created_at"`
} // @name Order

type Orders struct {
	Orders []*Order `json:"orders"`
} // @name Orders

type OrderStatus struct {
	Status string `json:"status"`
} // @name Orders

func toOrderResponse(order *domain.Order) *Order {
	return &Order{
		ID:        order.ID,
		Status:    order.Status,
		Total:     order.Total,
		CreatedAt: order.CreatedAt,
	}
}

func ToOrdersResponse(orders []*domain.Order) []*Order {
	ordersResponse := make([]*Order, len(orders))
	for _, order := range orders {
		ordersResponse = append(ordersResponse, toOrderResponse(order))
	}

	return ordersResponse
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
	items := make([]*OrderItemInfo, 0, len(orderInfo.Items))
	for _, item := range orderInfo.Items {
		items = append(items, toOrderItemResponse(item))
	}

	order := &OrderInfo{
		ID:        orderInfo.ID,
		Items:     items,
		Status:    orderInfo.Status,
		Total:     orderInfo.Total,
		CreatedAt: orderInfo.CreatedAt,
	}
	return order
}
