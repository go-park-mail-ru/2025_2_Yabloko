package domain

type Item struct {
	//Это ID из таблицы store_item
	ID          string
	Name        string
	Price       float64
	Description string
	CardImg     string
	TypesID     []string
}

type ItemType struct {
	ID   string
	Name string
}
