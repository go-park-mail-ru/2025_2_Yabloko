package domain

type Store struct {
	ID          string
	Name        string
	Description string
	CityID      string
	Address     string
	CardImg     string
	Rating      float64
	OpenAt      string
	ClosedAt    string
}

type StoreFilter struct {
	Limit  int
	LastID string
	Tag    string
	Sorted string
	Desc   bool
}
