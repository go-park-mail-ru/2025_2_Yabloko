package domain

type Item struct {
	ID      string   `json:"id"`       // store_item.id
	Name    string   `json:"name"`     // item.name
	Price   float64  `json:"price"`    // store_item.price
	TypesID []string `json:"types_id"` // массив type.id
	CardImg string   `json:"card_img"` // item.card_img
}

type ItemAgg struct {
	ID          string   `json:"id"`          // store_item.id
	Name        string   `json:"name"`        // item.name
	Price       float64  `json:"price"`       // store_item.price
	Description string   `json:"description"` // item.description
	CardImg     string   `json:"card_img"`    // item.card_img
	TypesID     []string `json:"types_id"`    // массив type.id
}

type ItemFilter struct {
	StoreID   string
	ItemTypes []string // фильтр ANY по type_id
	Sorted    string   // "name" или "price"
	Desc      bool
}

type ItemType struct {
	ID   string `json:"id"`   // type.id
	Name string `json:"name"` // type.name
}
