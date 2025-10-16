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

type StoreTag struct {
	ID   string
	Name string
}

type City struct {
	ID   string
	Name string
}

type StoreFilter struct {
	Limit  int
	LastID string
	TagID  string
	CityID string
	Sorted string
	Desc   bool
}
