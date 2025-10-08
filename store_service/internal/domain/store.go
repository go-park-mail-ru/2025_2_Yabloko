package domain

import "context"

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

type StoreRepository interface {
	GetStores(ctx context.Context, filter *StoreFilter) ([]*Store, error)
	GetStore(ctx context.Context, id string) (*Store, error)
	CreateStore(ctx context.Context, store *Store) error
}
