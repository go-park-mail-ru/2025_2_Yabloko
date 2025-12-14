package domain

type Store struct {
	ID          string
	Name        string
	Description string
	CityID      string
	Address     string
	CardImg     string
	Rating      float64
	TagID       string
	OpenAt      string
	ClosedAt    string
}

type StoreAgg struct {
	ID           string
	Name         string
	Description  string
	CityID       string
	Address      string
	CardImg      string
	Rating       float64
	TagsID       []string
	CategoriesID []string
	OpenAt       string
	ClosedAt     string
}

type StoreTag struct {
	ID   string
	Name string
}

type Category struct {
	ID   string
	Name string
}

type City struct {
	ID   string
	Name string
}

type StoreFilter struct {
	Limit       int
	Search      string
	LastID      string
	TagIDs      []string
	CategoryIDs []string
	CityID      string
	Sorted      string
	Desc        bool
}

type StoreReview struct {
	UserName  string
	Rating    float64
	Comment   string
	CreatedAt string
}

type CreateReviewRequest struct {
	UserID  string  `json:"user_id"`
	Rating  float64 `json:"rating"`
	Comment string  `json:"comment"`
}

type StoreSearchFilter struct {
	Search      string
	TagIDs      []string
	CategoryIDs []string
	CityID      string
	ItemTypes   []string
	MinPrice    float64
	MaxPrice    float64
	Limit       int
	LastID      string
}

type StoreWithItems struct {
	ID           string
	Name         string
	Description  string
	CityID       string
	Address      string
	CardImg      string
	Rating       float64
	TagsID       []string
	CategoriesID []string
	OpenAt       string
	ClosedAt     string
	Items        []*Item
}
