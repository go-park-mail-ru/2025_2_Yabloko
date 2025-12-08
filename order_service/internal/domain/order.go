package domain

import "time"

type OrderItemInfo struct {
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
	StoreID   string
	StoreName string
}

type Order struct {
	ID        string
	Status    string
	Total     float64
	CreatedAt time.Time
	StoreID   string
	StoreName string
}

type OrderFilter struct {
	UserID string
	Limit  int
	LastID string
	Status string
	Desc   bool
}
