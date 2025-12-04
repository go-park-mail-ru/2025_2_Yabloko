package domain

type Item struct {
	ID      string   `json:"id"`       // store_item.id
	Name    string   `json:"name"`     // item.name
	Price   float64  `json:"price"`    // store_item.price
	TypesID []string `json:"types_id"` // массив type.id
	CardImg string   `json:"card_img"` // item.card_img
}

// ItemAgg - агрегированный товар для GetItems by store
// ID здесь это store_item.id
type ItemAgg struct {
	ID          string   `json:"id"`          // store_item.id
	Name        string   `json:"name"`        // item.name
	Price       float64  `json:"price"`       // store_item.price
	Description string   `json:"description"` // item.description
	CardImg     string   `json:"card_img"`    // item.card_img
	TypesID     []string `json:"types_id"`    // массив type.id
}

type ItemType struct {
	ID   string `json:"id"`   // type.id
	Name string `json:"name"` // type.name
}
