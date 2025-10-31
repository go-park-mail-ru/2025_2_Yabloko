package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrderRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewOrderRepoPostgres(db PgxIface, log *logger.Logger) *OrderRepoPostgres {
	return &OrderRepoPostgres{
		db:  db,
		log: log,
	}
}

//go:embed sql/order/get_user_id.sql
var getOrderUser string

func (r *OrderRepoPostgres) GetOrderUserID(ctx context.Context, orderID string) (string, error) {
	r.log.Debug(ctx, "GetOrderUserID начало обработки", map[string]interface{}{})

	var userID string
	err := r.db.QueryRow(ctx, getOrderUser, orderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error(ctx, "GetOrderUserID заказ не существует", map[string]interface{}{"err": err, "id": orderID})
			return "", domain.ErrRowsNotFound
		}
		r.log.Error(ctx, "GetOrderUserID ошибка выполнения запроса", map[string]interface{}{"err": err, "id": orderID})
		return "", err
	}

	r.log.Debug(ctx, "GetOrderUserID завершено успешно", map[string]interface{}{})
	return userID, nil
}

//go:embed sql/order/insert_empty.sql
var insertEmptyOrder string

//go:embed sql/order/insert_item.sql
var insertItemOrder string

//go:embed sql/order/update_total.sql
var updateOrderTotal string

func (r *OrderRepoPostgres) CreateOrder(ctx context.Context, userID string) (string, error) {
	r.log.Debug(ctx, "CreateOrder начало обработки", map[string]interface{}{})

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// создаем пустую запись
	orderID := uuid.New().String()
	_, err = tx.Exec(ctx, insertEmptyOrder, orderID, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка создания заказа", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// переносим корзину
	_, err = tx.Exec(ctx, insertItemOrder, orderID, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка переноса товаров", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// сумма заказа
	_, err = tx.Exec(ctx, updateOrderTotal, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка обновления суммы", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	// очистка корзины
	_, err = tx.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка удаления записи", map[string]interface{}{"err": err, "id": userID})
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error(ctx, "CreateOrder ошибка закрытия транзакции", map[string]interface{}{"err": err, "id": userID})
		return "", domain.ErrInternalServer
	}

	r.log.Debug(ctx, "CreateOrder завершено успешно", map[string]interface{}{})
	return orderID, nil
}

//go:embed sql/order/update_status.sql
var updateOrderStatus string

func (r *OrderRepoPostgres) UpdateOrderStatus(ctx context.Context, id, status string) error {
	r.log.Debug(ctx, "UpdateOrderStatus начало обработки", map[string]interface{}{})

	_, err := r.db.Exec(ctx, updateOrderStatus, id, status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn(ctx, "UpdateOrderStatus заказ отсутствует",
				map[string]interface{}{"err": err, "id": id, "status": status})
			return domain.ErrRowsNotFound
		}
		r.log.Error(ctx, "UpdateOrderStatus ошибка обновления статуса",
			map[string]interface{}{"err": err, "id": id, "status": status})
		return err
	}

	r.log.Debug(ctx, "UpdateOrderStatus завершено успешно", map[string]interface{}{})
	return nil
}

//go:embed sql/order/get_order.sql
var getOrder string

func (r *OrderRepoPostgres) GetOrder(ctx context.Context, id string) (*domain.OrderInfo, error) {
	r.log.Debug(ctx, "GetOrder начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getOrder, id)
	if err != nil {
		r.log.Error(ctx, "GetOrder ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var order domain.OrderInfo
	var items []*domain.OrderItemInfo
	for rows.Next() {
		var item domain.OrderItemInfo
		err = rows.Scan(
			&order.ID,
			&order.Total,
			&order.Status,
			&order.CreatedAt,
			&item.ID,
			&item.Name,
			&item.CardImg,
			&item.Price,
			&item.Quantity,
		)
		if err != nil {
			r.log.Error(ctx, "GetOrder ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		items = append(items, &item)
	}

	err = rows.Err()
	if err != nil {
		r.log.Error(ctx, "GetOrder ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn(ctx, "GetOrder пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	order.Items = items
	r.log.Debug(ctx, "GetOrder завершено успешно", map[string]interface{}{})
	return &order, nil
}

//go:embed sql/order/get_user_orders.sql
var getUserOrders string

func (r *OrderRepoPostgres) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	r.log.Debug(ctx, "GetOrdersUser начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getUserOrders, userID)
	if err != nil {
		r.log.Error(ctx, "GetOrdersUser ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		err = rows.Scan(
			&order.ID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
		)
		if err != nil {
			r.log.Error(ctx, "GetOrdersUser ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		orders = append(orders, &order)
	}

	err = rows.Err()
	if err != nil {
		r.log.Error(ctx, "GetOrdersUser ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}

	if len(orders) == 0 {
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetOrdersUser завершено успешно", map[string]interface{}{})
	return orders, nil
}
