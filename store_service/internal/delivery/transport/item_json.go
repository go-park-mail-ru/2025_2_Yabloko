package transport

type Item struct {
	// ID из таблицы store_item
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Description string   `json:"description"`
	CardImg     string   `json:"card_img"`
	TypesID     []string `json:"types_id"`
} // @name Item

type ItemType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
} // @name ItemType
