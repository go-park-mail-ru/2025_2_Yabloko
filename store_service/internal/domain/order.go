package domain

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
	CreatedAt string
}

type Order struct {
	ID        string
	Status    string
	Total     float64
	CreatedAt string
}
