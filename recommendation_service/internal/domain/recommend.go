package domain

type RecommendedItem struct {
	ID      string
	StoreID string
	Name    string
	Price   float64
	CardImg string
	Score   float64
}

type HomeRecommendFilter struct {
	UserID string
	Limit  int
}
