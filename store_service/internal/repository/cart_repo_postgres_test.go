package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestCartRepoPostgres_GetCartItems(t *testing.T) {
	type testCase struct {
		name          string
		id            string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.CartItem
		expectedError error
	}

	query := `
		select
		    c.id as cart_id,
		    si.id as id,
			it.name as name,
			it.card_img as card_img,
			si.price as price,
			ci.quantity as quantity
		from cart c
		join cart_item ci on ci.cart_id = c.id
		join store_item si on si.id = ci.store_item_id
		join item it on it.id = si.item_id
		where c.user_id = \$1
		order by ci.created_at;
	`

	uid1 := "00000000-0000-0000-0000-000000000001"
	uid2 := "00000000-0000-0000-0000-000000000002"
	name1 := "name1"
	name2 := "name2"
	cartImg1 := "cartImg1"
	cartImg2 := "cartImg2"
	price1 := 1.2
	price2 := 2.2
	quantity1 := 3
	quantity2 := 4

	item1 := &domain.CartItem{
		ID:       uid1,
		Name:     name1,
		CardImg:  cartImg1,
		Price:    price1,
		Quantity: quantity1,
	}
	item2 := &domain.CartItem{
		ID:       uid2,
		Name:     name2,
		CardImg:  cartImg2,
		Price:    price2,
		Quantity: quantity2,
	}

	tests := []testCase{
		{
			name: "успешный запрос больше 1 элемента",
			id:   uid1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"cart_id", "id", "name", "card_img", "price", "quantity"}).
					AddRow(uid1, uid1, name1, cartImg1, price1, quantity1).
					AddRow(uid1, uid2, name2, cartImg2, price2, quantity2)
				mock.ExpectQuery(query).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.CartItem{item1, item2},
			expectedError: nil,
		},
		{
			name: "успешный запрос 1 элемент",
			id:   uid1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"cart_id", "id", "name", "card_img", "price", "quantity"}).
					AddRow(uid1, uid1, name1, cartImg1, price1, quantity1)
				mock.ExpectQuery(query).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.CartItem{item1},
			expectedError: nil,
		},
		{
			name: "пустой ответ",
			id:   uid1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"cart_id", "id", "name", "card_img", "price", "quantity"})
				mock.ExpectQuery(query).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка запроса",
			id:   uid1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(uid1).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "ошибка при чтении строки",
			id:   uid1,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"cart_id", "id", "name", "card_img", "price", "quantity"}).
					AddRow(uid1, uid1, name1, cartImg1, price1, quantity1).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(query).
					WithArgs(uid1).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewCartRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetCartItems(context.Background(), tt.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestCartRepoPostgres_DeleteCartItems(t *testing.T) {
	type testCase struct {
		name          string
		id            string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}

	query := `
		delete from cart_item
		where cart_id = \(select id from cart where user_id = \$1\)
		returning cart_id
	`

	userID := "00000000-0000-0000-0000-000000000111"

	tests := []testCase{
		{
			name: "успешное удаление",
			id:   userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(query).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: nil,
		},
		{
			name: "ошибка базы данных",
			id:   userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(query).
					WithArgs(userID).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewCartRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			err = repo.DeleteCartItems(context.Background(), tt.id)

			require.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCartRepoPostgres_UpdateCartItems(t *testing.T) {
	type testCase struct {
		name          string
		id            string
		newItems      *domain.CartUpdate
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}

	deleteQuery := `
		delete from cart_item
		where cart_id = \(select id from cart where user_id = \$1\)
		returning cart_id
	`
	insertQuery := `
		insert into cart_item \(id, cart_id, store_item_id, quantity\)
		values \(\$1, \$2, \$3, \$4\);
	`

	userID := "00000000-0000-0000-0000-000000000111"
	cartID := "00000000-0000-0000-0000-000000000222"

	item1 := &domain.ItemUpdate{
		ID:       "00000000-0000-0000-0000-000000000121",
		Quantity: 2,
	}
	item2 := &domain.ItemUpdate{
		ID:       "00000000-0000-0000-0000-000000000112",
		Quantity: 5,
	}

	update := &domain.CartUpdate{
		Items: []*domain.ItemUpdate{item1, item2},
	}

	tests := []testCase{
		{
			name:     "успешное обновление корзины",
			id:       userID,
			newItems: update,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				mock.ExpectQuery(deleteQuery).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow(cartID))

				mock.ExpectExec(insertQuery).
					WithArgs(pgxmock.AnyArg(), cartID, item1.ID, item1.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertQuery).
					WithArgs(pgxmock.AnyArg(), cartID, item2.ID, item2.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:     "ошибка при удалении",
			id:       userID,
			newItems: update,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(deleteQuery).
					WithArgs(userID).
					WillReturnError(domain.ErrInternalServer)
				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
		},
		{
			name:     "ошибка при вставке",
			id:       userID,
			newItems: update,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(deleteQuery).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow(cartID))

				mock.ExpectExec(insertQuery).
					WithArgs(pgxmock.AnyArg(), cartID, item1.ID, item1.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertQuery).
					WithArgs(pgxmock.AnyArg(), cartID, item2.ID, item2.Quantity).
					WillReturnError(domain.ErrInternalServer)

				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
		},
		{
			name:     "ошибка при начале транзакции",
			id:       cartID,
			newItems: update,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewCartRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			err = repo.UpdateCartItems(context.Background(), tt.id, tt.newItems)

			require.Equal(t, tt.expectedError, err)
		})
	}
}
