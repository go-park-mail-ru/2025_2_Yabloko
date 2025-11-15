package domain

import "time"

type OrderItemInfo struct {
	// id - store_item_id
	ID       string
	Name     string
	CardImg  string
	Price    float64
	Quantity int
}

type OrderInfo struct {
	ID        string
	Items     []*OrderItemInfo
	Status    string
	Total     float64
	CreatedAt time.Time
}

type Order struct {
	ID        string
	Status    string
	Total     float64
	CreatedAt time.Time
}

type OrderFilter struct {
	UserID string
	Limit  int
	LastID string
	Status string
	Desc   bool // сортировка по убыванию (новые сначала)
}
