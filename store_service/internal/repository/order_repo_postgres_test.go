package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestOrderRepoPostgres_GetOrder(t *testing.T) {
	type testCase struct {
		name          string
		id            string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   *domain.OrderInfo
		expectedError error
	}

	query := `
		select
			o.id as order_id,
			o.total_price as total,
			o.status as status,
			o.created_at as created_at,
			si.id as store_item_id,
			i.name as name,
			i.card_img as card_img,
			oi.price as price,
			oi.quantity as quantity
		from "order" o
		join order_item oi on oi.order_id = o.id
		join store_item si on si.id = oi.store_item_id
		join item i on i.id = si.item_id
		where o.id = \$1
		order by oi.created_at;
	`

	orderID := "00000000-0000-0000-0000-000000000001"
	storeItemID1 := "00000000-0000-0000-0000-000000000002"
	storeItemID2 := "00000000-0000-0000-0000-000000000003"

	item1 := &domain.OrderItemInfo{
		ID:       storeItemID1,
		Name:     "item1",
		CardImg:  "img1.png",
		Price:    10.5,
		Quantity: 2,
	}
	item2 := &domain.OrderItemInfo{
		ID:       storeItemID2,
		Name:     "item2",
		CardImg:  "img2.png",
		Price:    20.0,
		Quantity: 1,
	}

	order := &domain.OrderInfo{
		ID:        orderID,
		Total:     41.0,
		Status:    "paid",
		CreatedAt: "2024-01-01",
		Items:     []*domain.OrderItemInfo{item1, item2},
	}

	tests := []testCase{
		{
			name: "успешный запрос",
			id:   orderID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"order_id", "total", "status", "created_at",
					"store_item_id", "name", "card_img", "price", "quantity",
				}).
					AddRow(orderID, order.Total, order.Status, order.CreatedAt,
						item1.ID, item1.Name, item1.CardImg, item1.Price, item1.Quantity).
					AddRow(orderID, order.Total, order.Status, order.CreatedAt,
						item2.ID, item2.Name, item2.CardImg, item2.Price, item2.Quantity)

				mock.ExpectQuery(query).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			expectedRes:   order,
			expectedError: nil,
		},
		{
			name: "пустой ответ",
			id:   orderID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"order_id", "total", "status", "created_at",
					"store_item_id", "name", "card_img", "price", "quantity",
				})
				mock.ExpectQuery(query).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name: "ошибка запроса к БД",
			id:   orderID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(orderID).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "ошибка при сканировании строки",
			id:   orderID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"order_id", "total", "status", "created_at",
					"store_item_id", "name", "card_img", "price", "quantity",
				}).AddRow(orderID, order.Total, order.Status, order.CreatedAt,
					item1.ID, item1.Name, item1.CardImg, item1.Price, item1.Quantity).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(query).
					WithArgs(orderID).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name: "ошибка после чтения строк",
			id:   orderID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"order_id", "total", "status", "created_at",
					"store_item_id", "name", "card_img", "price", "quantity",
				}).AddRow(orderID, order.Total, order.Status, order.CreatedAt,
					item1.ID, item1.Name, item1.CardImg, item1.Price, item1.Quantity).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(query).
					WithArgs(orderID).
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
			repo := NewOrderRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetOrder(context.Background(), tt.id)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestOrderRepoPostgres_GetOrdersUser(t *testing.T) {
	type testCase struct {
		name          string
		userID        string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedRes   []*domain.Order
		expectedError error
	}

	query := `
		select
			o.id as id,
			o.status as status,
			o.total_price as total,
			o.created_at as created_at
		from "order" o
		where o.user_id = \$1
		order by o.created_at;
	`

	userID := "00000000-0000-0000-0000-000000000123"

	order1 := &domain.Order{
		ID:        "11111111-1111-1111-1111-111111111111",
		Status:    "paid",
		Total:     50.0,
		CreatedAt: "2024-01-01",
	}
	order2 := &domain.Order{
		ID:        "22222222-2222-2222-2222-222222222222",
		Status:    "shipped",
		Total:     100.0,
		CreatedAt: "2024-01-01",
	}

	tests := []testCase{
		{
			name:   "успешное получение заказов пользователя",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "status", "total", "created_at"}).
					AddRow(order1.ID, order1.Status, order1.Total, order1.CreatedAt).
					AddRow(order2.ID, order2.Status, order2.Total, order2.CreatedAt)

				mock.ExpectQuery(query).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedRes:   []*domain.Order{order1, order2},
			expectedError: nil,
		},
		{
			name:   "ошибка при выполнении запроса к БД",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(userID).
					WillReturnError(domain.ErrInternalServer)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name:   "ошибка при сканировании строки",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "status", "total", "created_at"}).
					AddRow(order1.ID, order1.Status, order1.Total, order1.CreatedAt).
					RowError(0, domain.ErrInternalServer)

				mock.ExpectQuery(query).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrInternalServer,
		},
		{
			name:   "пустой результат",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "status", "total", "created_at"})
				mock.ExpectQuery(query).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedRes:   nil,
			expectedError: domain.ErrRowsNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewOrderRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			res, err := repo.GetOrdersUser(context.Background(), tt.userID)

			require.Equal(t, tt.expectedError, err)
			require.Equal(t, tt.expectedRes, res)
		})
	}
}

func TestOrderRepoPostgres_UpdateOrderStatus(t *testing.T) {
	type testCase struct {
		name          string
		id            string
		status        string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
	}

	query := `
		update "order"
		set status = \$2
		where id = \$1
		returning id
	`

	orderID := "00000000-0000-0000-0000-000000000123"
	status := "pending"

	tests := []testCase{
		{
			name:   "успешное обновление статуса",
			id:     orderID,
			status: status,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(orderID, status).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(orderID))
			},
			expectedError: nil,
		},
		{
			name:   "пустой ответ",
			id:     orderID,
			status: status,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(orderID, status).
					WillReturnRows(pgxmock.NewRows([]string{"id"}))
			},
			expectedError: domain.ErrRowsNotFound,
		},
		{
			name:   "ошибка запроса к БД",
			id:     orderID,
			status: status,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(query).
					WithArgs(orderID, status).
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
			repo := NewOrderRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			err = repo.UpdateOrderStatus(context.Background(), tt.id, tt.status)

			require.Equal(t, tt.expectedError, err)
			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestOrderRepoPostgres_CreateOrder(t *testing.T) {
	type testCase struct {
		name          string
		userID        string
		mockSetup     func(mock pgxmock.PgxPoolIface)
		expectedError error
		expectID      bool
	}

	insertOrderQuery := `
		insert into "order" \(id, user_id, total_price\)
		values \(\$1, \$2, 0\);
	`
	insertItemsQuery := `
		insert into order_item \(id, order_id, store_item_id, price, quantity\)
		select gen_random_uuid\(\), \$1, si.id, si.price, ci.quantity
		from cart_item ci
		join cart c on c.id = ci.cart_id
		join store_item si on si.id = ci.store_item_id
		where c.user_id = \$2;
	`
	updateTotalQuery := `
		update "order"
		set total_price = \(
			select sum\(si.price \* ci.quantity\)
			from cart_item ci
			join cart c ON c.id = ci.cart_id
			join store_item si ON si.id = ci.store_item_id
			where c.user_id = \$1
		\)
		where id = \$2;
	`
	deleteCartQuery := `
		delete from cart_item
		where cart_id = \(select id from cart where user_id = \$1\);
	`

	userID := "00000000-0000-0000-0000-000000000123"

	tests := []testCase{
		{
			name:   "успешное создание заказа",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(insertItemsQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))

				mock.ExpectExec(updateTotalQuery).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				mock.ExpectExec(deleteCartQuery).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))

				mock.ExpectCommit()
			},
			expectedError: nil,
			expectID:      true,
		},
		{
			name:   "ошибка при начале транзакции",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
		{
			name:   "ошибка при insert order",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnError(domain.ErrInternalServer)
				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
		{
			name:   "ошибка при insert order_item",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertItemsQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnError(domain.ErrInternalServer)
				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
		{
			name:   "ошибка при обновлении суммы заказа",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertItemsQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mock.ExpectExec(updateTotalQuery).
					WithArgs(userID).
					WillReturnError(domain.ErrInternalServer)
				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
		{
			name:   "ошибка при удалении корзины",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertItemsQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mock.ExpectExec(updateTotalQuery).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec(deleteCartQuery).
					WithArgs(userID).
					WillReturnError(domain.ErrInternalServer)
				mock.ExpectRollback()
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
		{
			name:   "ошибка при завершении транзакции",
			userID: userID,
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(insertOrderQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(insertItemsQuery).
					WithArgs(pgxmock.AnyArg(), userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mock.ExpectExec(updateTotalQuery).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec(deleteCartQuery).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))
				mock.ExpectCommit().WillReturnError(domain.ErrInternalServer)
			},
			expectedError: domain.ErrInternalServer,
			expectID:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			log := logger.NewNilLogger()
			repo := NewOrderRepoPostgres(mockPool, log)

			tt.mockSetup(mockPool)

			ID, err := repo.CreateOrder(context.Background(), tt.userID)

			require.Equal(t, tt.expectedError, err)
			if tt.expectID {
				require.NotEmpty(t, ID)
			} else {
				require.Empty(t, ID)
			}
		})
	}
}
