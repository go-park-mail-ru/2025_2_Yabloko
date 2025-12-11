package domain

import "time"

type OrderItemInfo struct {
	ID       string
	Name     string
	CardImg  string
	Price    float64
	Quantity int
}

type StoreInfo struct {
	ID      string
	Name    string
	CardImg string
	Items   []*OrderItemInfo
}

type OrderInfo struct {
	ID        string
	Stores    []*StoreInfo
	Status    string
	Total     float64
	CreatedAt time.Time
}

type Order struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	StoreID   string    `json:"store_id"`
	StoreName string    `json:"store_name"`
}

type OrderFilter struct {
	UserID string
	Limit  int
	LastID string
	Status string
	Desc   bool
}
