package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestStoreRepoPostgres_GetStore(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		storeID       string
		expectedRes   *domain.Store
		expectedError error
	}

	storeID := "00000000-0000-0000-0000-000000000001"
	store := &domain.Store{
		ID:          storeID,
		Name:        "Store1",
		Description: "Description1",
		CityID:      "City1",
		Address:     "Address1",
		CardImg:     "img1",
		Rating:      4.5,
		OpenAt:      "08:00",
		ClosedAt:    "22:00",
	}

	tests := []testCase{
		{
			name:    "успешный запрос",
			storeID: storeID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"}).
					AddRow(store.ID, store.Name, store.Description, store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store where id = \$1`).
					WithArgs(storeID).
					WillReturnRows(rows)
			},
			expectedRes:   store,
			expectedError: nil,
		},
		{
			name:    "пустой результат",
			storeID: storeID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"})
				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store where id = \$1`).
					WithArgs(storeID).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name:    "ошибка запроса",
			storeID: storeID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store where id = \$1`).
					WithArgs(storeID).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewStoreRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetStore(context.Background(), tt.storeID)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestStoreRepoPostgres_CreateStore(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		inputStore    *domain.Store
		expectedError error
	}

	store := &domain.Store{
		Name:        "Store1",
		Description: "Description1",
		CityID:      "City1",
		Address:     "Address1",
		CardImg:     "img1",
		Rating:      4.5,
		OpenAt:      "08:00",
		ClosedAt:    "22:00",
	}

	tests := []testCase{
		{
			name:       "успешное создание",
			inputStore: store,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into store \(id, name, description, city_id, address, card_img, rating, open_at, closed_at\)
		values \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\)`).
					WithArgs(pgxmock.AnyArg(), store.Name, store.Description, store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
		{
			name:       "уникальный конфликт",
			inputStore: store,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into store \(id, name, description, city_id, address, card_img, rating, open_at, closed_at\)
		values \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\)`).
					WithArgs(pgxmock.AnyArg(), store.Name, store.Description, store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt).
					WillReturnError(errors.New("SQLSTATE 23505"))
			},
			expectedError: domain.ErrStoreExist,
		},
		{
			name:       "другая ошибка бд",
			inputStore: store,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`insert into store \(id, name, description, city_id, address, card_img, rating, open_at, closed_at\)
		values \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9\)`).
					WithArgs(pgxmock.AnyArg(), store.Name, store.Description, store.CityID, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewStoreRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			err = repo.CreateStore(context.Background(), tt.inputStore)
			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestStoreRepoPostgres_GetStores(t *testing.T) {
	type testCase struct {
		name          string
		filter        *domain.StoreFilter
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.Store
		expectedError error
	}

	store1 := &domain.Store{
		ID:          "00000000-0000-0000-0000-000000000001",
		Name:        "Store1",
		Description: "Desc1",
		CityID:      "City1",
		Address:     "Addr1",
		CardImg:     "Img1",
		Rating:      4.5,
		OpenAt:      "08:00",
		ClosedAt:    "22:00",
	}

	store2 := &domain.Store{
		ID:          "00000000-0000-0000-0000-000000000002",
		Name:        "Store2",
		Description: "Desc2",
		CityID:      "City2",
		Address:     "Addr2",
		CardImg:     "Img2",
		Rating:      3.8,
		OpenAt:      "09:00",
		ClosedAt:    "21:00",
	}

	tests := []testCase{
		{
			name:   "без фильтров",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"}).
					AddRow(store1.ID, store1.Name, store1.Description, store1.CityID, store1.Address, store1.CardImg, store1.Rating, store1.OpenAt, store1.ClosedAt).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt)

				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store order by id limit \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store1, store2},
			expectedError: nil,
		},
		{
			name:   "с фильтром по id",
			filter: &domain.StoreFilter{Limit: 5, LastID: "00000000-0000-0000-0000-000000000001"},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"}).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt)

				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store where id > \$1 order by id limit \$2`).
					WithArgs("00000000-0000-0000-0000-000000000001", 5).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store2},
			expectedError: nil,
		},
		{
			name:   "пустой результат",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"})
				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store order by id limit \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name:   "ошибка запроса",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store order by id limit \$1`).
					WithArgs(10).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   nil,
			expectedError: errors.New("db error"),
		},
		{
			name:   "ошибка при чтении",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at"}).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(`select id, name, description, city_id, address, card_img, rating, open_at, closed_at from store order by id limit \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewStoreRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetStores(context.Background(), tt.filter)
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}
