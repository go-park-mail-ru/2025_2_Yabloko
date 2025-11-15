package domain

type CartItem struct {
	// id - store_item_id
	ID       string
	Name     string
	CardImg  string
	Price    float64
	Quantity int
}

type Cart struct {
	Items []*CartItem
}

type ItemUpdate struct {
	// id - store_item_id
	ID       string
	Quantity int
}

type CartUpdate struct {
	Items []*ItemUpdate
}
