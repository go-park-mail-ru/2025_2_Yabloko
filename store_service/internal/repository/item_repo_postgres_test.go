package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"

	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestItemRepoPostgres_GetItemTypes(t *testing.T) {

	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.ItemType
		expectedError error
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"

	tests := []testCase{
		{
			name: "успешный запрос",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(uid1, name1).
					AddRow(uid2, name2)

				mock.ExpectQuery(`select type.id, type.name
		from store_item join item_type on store_item.item_id = item_type.item_id
		join type on store_item.type_id = type.id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes: []*domain.ItemType{
				{ID: uid1, Name: name1},
				{ID: uid2, Name: name2},
			},
			expectedError: nil,
		},
		{
			name: "ошибка при запросе",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select type.id, type.name
		from store_item join item_type on store_item.item_id = item_type.item_id
		join type on store_item.type_id = type.id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "пустой результат",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"})
				mock.ExpectQuery(`select type.id, type.name
		from store_item join item_type on store_item.item_id = item_type.item_id
		join type on store_item.type_id = type.id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка при чтении",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(uid1, name1).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(`select type.id, type.name
		from store_item join item_type on store_item.item_id = item_type.item_id
		join type on store_item.type_id = type.id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				panic(err)
			}
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewItemRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetItemTypes(context.Background(), uid1)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestItemRepoPostgres_GetItems(t *testing.T) {

	type testCase struct {
		name          string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.Item
		expectedError error
	}

	uid1 := "00000000-0000-0000-0000-000000000001"
	name1 := "name1"
	description1 := "description1"
	price1 := 1.0
	cardImg1 := "card_img1"

	uid2 := "00000000-0000-0000-0000-000000000002"
	name2 := "name2"
	description2 := "description2"
	price2 := 2.0
	cardImg2 := "card_img2"

	tests := []testCase{
		{
			name: "успешный запрос",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "price", "description", "card_img", "type_id"}).
					AddRow(uid1, name1, price1, description1, cardImg1, uid1).
					AddRow(uid2, name2, price2, description2, cardImg2, uid2)

				mock.ExpectQuery(`select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes: []*domain.Item{
				{
					ID:          uid1,
					Name:        name1,
					Description: description1,
					Price:       price1,
					CardImg:     cardImg1,
					TypeID:      uid1,
				},
				{
					ID:          uid2,
					Name:        name2,
					Description: description2,
					Price:       price2,
					CardImg:     cardImg2,
					TypeID:      uid2,
				},
			},
			expectedError: nil,
		},
		{
			name: "успешный запрос несколько типов",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "price", "description", "card_img", "type_id"}).
					AddRow(uid1, name1, price1, description1, cardImg1, uid1).
					AddRow(uid1, name1, price1, description1, cardImg1, uid2).
					AddRow(uid2, name2, price2, description2, cardImg2, uid1).
					AddRow(uid2, name2, price2, description2, cardImg2, uid2)

				mock.ExpectQuery(`select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes: []*domain.Item{
				{
					ID:          uid1,
					Name:        name1,
					Description: description1,
					Price:       price1,
					CardImg:     cardImg1,
					TypeID:      uid1,
				},
				{
					ID:          uid1,
					Name:        name1,
					Description: description1,
					Price:       price1,
					CardImg:     cardImg1,
					TypeID:      uid2,
				},
				{
					ID:          uid2,
					Name:        name2,
					Description: description2,
					Price:       price2,
					CardImg:     cardImg2,
					TypeID:      uid1,
				},
				{
					ID:          uid2,
					Name:        name2,
					Description: description2,
					Price:       price2,
					CardImg:     cardImg2,
					TypeID:      uid2,
				},
			},
			expectedError: nil,
		},
		{
			name: "ошибка при запросе",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "пустой результат",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "price", "description", "card_img", "type_id"})
				mock.ExpectQuery(`select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка при чтении",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "price", "description", "card_img", "type_id"}).
					AddRow(uid1, name1, price1, description1, cardImg1, uid1).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(`select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = \$1`).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			if err != nil {
				panic(err)
			}
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewItemRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetItems(context.Background(), uid1)

			require.Equal(t, tt.expectedError, err)
			require.ElementsMatch(t, tt.expectedRes, res)
		})
	}
}
