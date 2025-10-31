package repository

import (
	"apple_backend/custom_errors"
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
		expectedRes   []*domain.Store
		expectedError error
	}

	storeID := "00000000-0000-0000-0000-000000000001"
	stores := []*domain.Store{
		{
			ID:          storeID,
			Name:        "Store1",
			Description: "Description1",
			CityID:      "City1",
			Address:     "Address1",
			CardImg:     "img1",
			Rating:      4.5,
			TagID:       storeID,
			OpenAt:      "08:00",
			ClosedAt:    "22:00",
		},
	}

	tests := []testCase{
		{
			name:    "успешный запрос",
			storeID: storeID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(stores[0].ID, stores[0].Name, stores[0].Description, stores[0].CityID, stores[0].Address, stores[0].CardImg, stores[0].Rating, stores[0].OpenAt, stores[0].ClosedAt, stores[0].TagID)
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id
		from store s left join store_tag st on st.store_id = s.id where id = \$1`).
					WithArgs(storeID).
					WillReturnRows(rows)
			},
			expectedRes:   stores,
			expectedError: nil,
		},
		{
			name:    "пустой результат",
			storeID: storeID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"})
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id
		from store s left join store_tag st on st.store_id = s.id where id = \$1`).
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
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id
		from store s left join store_tag st on st.store_id = s.id where id = \$1`).
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

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	description1 := "description1"
	address1 := "address1"
	rating1 := 1.0
	cardImg1 := "card_img1"
	openAt1 := "open_at1"
	closeAt1 := "close_at1"

	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"
	description2 := "description2"
	address2 := "address2"
	rating2 := 2.0
	cardImg2 := "card_img2"
	openAt2 := "open_at2"
	closeAt2 := "close_at2"

	store1 := &domain.Store{
		ID:          uid1,
		Name:        name1,
		Description: description1,
		CityID:      uid1,
		Address:     address1,
		CardImg:     cardImg1,
		Rating:      rating1,
		TagID:       uid1,
		OpenAt:      openAt1,
		ClosedAt:    closeAt1,
	}

	store2 := &domain.Store{
		ID:          uid2,
		Name:        name2,
		Description: description2,
		CityID:      uid2,
		Address:     address2,
		CardImg:     cardImg2,
		Rating:      rating2,
		TagID:       uid2,
		OpenAt:      openAt2,
		ClosedAt:    closeAt2,
	}

	tests := []testCase{
		{
			name:   "без фильтров",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(store1.ID, store1.Name, store1.Description, store1.CityID, store1.Address, store1.CardImg, store1.Rating, store1.OpenAt, store1.ClosedAt, store1.TagID).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt, store2.TagID)
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id order by s.id limit \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store1, store2},
			expectedError: nil,
		},
		{
			name:   "с фильтром по id",
			filter: &domain.StoreFilter{Limit: 5, LastID: uid1},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt, store2.TagID)

				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id where s.id > \$1 order by s.id limit \$2`).
					WithArgs(uid1, 5).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store2},
			expectedError: nil,
		},
		{
			name:   "с фильтром по тегу и городу",
			filter: &domain.StoreFilter{Limit: 10, TagID: uid1, CityID: uid1},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(store1.ID, store1.Name, store1.Description, store1.CityID, store1.Address, store1.CardImg, store1.Rating, store1.OpenAt, store1.ClosedAt, store1.TagID)

				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id where exists \(select 1 from store_tag st2 where st2.store_id = s.id and st2.tag_id = \$1\) and s.city_id = \$2 order by s.id limit \$3`).
					WithArgs(uid1, uid1, 10).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store1},
			expectedError: nil,
		},
		{
			name:   "с сортировкой по рейтингу desc",
			filter: &domain.StoreFilter{Limit: 10, Sorted: "rating", Desc: true},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt, store2.TagID).
					AddRow(store1.ID, store1.Name, store1.Description, store1.CityID, store1.Address, store1.CardImg, store1.Rating, store1.OpenAt, store1.ClosedAt, store1.TagID)

				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id order by s.rating desc, s.id limit \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Store{store2, store1},
			expectedError: nil,
		},
		{
			name:   "пустой результат",
			filter: &domain.StoreFilter{Limit: 10},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"})
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id order by s.id limit \$1`).
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
				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id order by s.id limit \$1`).
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
				rows := pgxmock.NewRows([]string{"id", "name", "description", "city_id", "address", "card_img", "rating", "open_at", "closed_at", "tag_id"}).
					AddRow(store2.ID, store2.Name, store2.Description, store2.CityID, store2.Address, store2.CardImg, store2.Rating, store2.OpenAt, store2.ClosedAt, store2.TagID).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(`select s.id, s.name, s.description, s.city_id, s.address, s.card_img, s.rating, s.open_at, s.closed_at, st.tag_id from store s left join store_tag st on st.store_id = s.id order by s.id limit \$1`).
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

func TestStoreRepoPostgres_GetCities(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.City
		expectedError error
	}

	city1 := &domain.City{
		ID:   "00000000-0000-0000-0000-000000000001",
		Name: "City1",
	}
	city2 := &domain.City{
		ID:   "00000000-0000-0000-0000-000000000002",
		Name: "City2",
	}

	tests := []testCase{
		{
			name: "успешный запрос",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(city1.ID, city1.Name).
					AddRow(city2.ID, city2.Name)

				mock.ExpectQuery(`select id, name from city`).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.City{city1, city2},
			expectedError: nil,
		},
		{
			name: "пустой результат",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery(`select id, name from city`).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка запроса",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, name from city`).
					WillReturnError(custom_errors.InnerErr)
			},
			expectedRes:   nil,
			expectedError: custom_errors.InnerErr,
		},
		{
			name: "ошибка при чтении строки",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(city1.ID, city1.Name).
					RowError(0, custom_errors.InnerErr)

				mock.ExpectQuery(`select id, name from city`).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: custom_errors.InnerErr,
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

			res, err := repo.GetCities(context.Background())
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestStoreRepoPostgres_GetTags(t *testing.T) {
	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.StoreTag
		expectedError error
	}

	tag1 := &domain.StoreTag{
		ID:   "00000000-0000-0000-0000-000000000001",
		Name: "tag1",
	}
	tag2 := &domain.StoreTag{
		ID:   "00000000-0000-0000-0000-000000000002",
		Name: "tag2",
	}

	tests := []testCase{
		{
			name: "успешный запрос",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(tag1.ID, tag1.Name).
					AddRow(tag2.ID, tag2.Name)

				mock.ExpectQuery(`select id, name from tag`).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.StoreTag{tag1, tag2},
			expectedError: nil,
		},
		{
			name: "пустой результат",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery(`select id, name from tag`).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка запроса",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select id, name from tag`).
					WillReturnError(custom_errors.InnerErr)
			},
			expectedRes:   nil,
			expectedError: custom_errors.InnerErr,
		},
		{
			name: "ошибка при чтении строки",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(tag1.ID, tag1.Name).
					RowError(0, custom_errors.InnerErr)

				mock.ExpectQuery(`select id, name from tag`).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: custom_errors.InnerErr,
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

			res, err := repo.GetTags(context.Background())
			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}
