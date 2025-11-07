package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed sql/order/get_user_id.sql
var getOrderUser string

//go:embed sql/order/insert_empty.sql
var insertEmptyOrder string

//go:embed sql/order/insert_item.sql
var insertItemOrder string

//go:embed sql/order/update_total.sql
var updateOrderTotal string

//go:embed sql/order/update_status.sql
var updateOrderStatus string

//go:embed sql/order/get_order.sql
var getOrder string

//go:embed sql/order/get_user_orders.sql
var getUserOrders string

// Уже объявлено в cart_repo.go
//
//go:embed sql/cart/delete_items.sql
var deleteCartItemsForOrder string

type OrderRepoPostgres struct {
	db  PgxIface
	log logger.Logger
}

func NewOrderRepoPostgres(db PgxIface, log logger.Logger) *OrderRepoPostgres {
	return &OrderRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *OrderRepoPostgres) GetOrderUserID(ctx context.Context, orderID string) (string, error) {
	r.log.Debug("GetOrderUserID начало обработки", slog.String("order_id", orderID))

	var userID string
	err := r.db.QueryRow(ctx, getOrderUser, orderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn("GetOrderUserID заказ не существует", slog.String("order_id", orderID))
			return "", domain.ErrRowsNotFound
		}
		r.log.Error("GetOrderUserID ошибка выполнения запроса", slog.String("order_id", orderID), slog.Any("err", err))
		return "", err
	}

	r.log.Debug("GetOrderUserID завершено успешно", slog.String("order_id", orderID))
	return userID, nil
}

func (r *OrderRepoPostgres) CreateOrder(ctx context.Context, userID string) (string, error) {
	r.log.Debug("CreateOrder начало обработки", slog.String("user_id", userID))

	// Проверка: есть ли товары в корзине?
	var cnt int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM cart_item WHERE user_id = $1", userID).Scan(&cnt)
	if err != nil {
		r.log.Error("CreateOrder ошибка проверки корзины", slog.String("user_id", userID), slog.Any("err", err))
		return "", err
	}
	if cnt == 0 {
		r.log.Warn("CreateOrder корзина пуста", slog.String("user_id", userID))
		return "", domain.ErrCartEmpty
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error("CreateOrder ошибка начала транзакции", slog.String("user_id", userID), slog.Any("err", err))
		return "", err
	}
	defer tx.Rollback(ctx)

	// 1. Создаём заказ
	orderID := uuid.New().String()
	_, err = tx.Exec(ctx, insertEmptyOrder, orderID, userID)
	if err != nil {
		r.log.Error("CreateOrder ошибка создания заказа", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", err
	}

	// 2. Переносим товары из корзины
	_, err = tx.Exec(ctx, insertItemOrder, orderID, userID)
	if err != nil {
		r.log.Error("CreateOrder ошибка переноса товаров", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", err
	}

	// 3. Обновляем итоговую сумму
	_, err = tx.Exec(ctx, updateOrderTotal, orderID)
	if err != nil {
		r.log.Error("CreateOrder ошибка обновления суммы", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", err
	}

	// 4. Очищаем корзину
	_, err = tx.Exec(ctx, deleteCartItemsForOrder, userID)
	if err != nil {
		r.log.Error("CreateOrder ошибка удаления записи", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error("CreateOrder ошибка закрытия транзакции", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	r.log.Debug("CreateOrder завершено успешно", slog.String("user_id", userID), slog.String("order_id", orderID))
	return orderID, nil
}

func (r *OrderRepoPostgres) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	r.log.Debug("UpdateOrderStatus начало обработки", slog.String("order_id", orderID), slog.String("status", status))

	_, err := r.db.Exec(ctx, updateOrderStatus, orderID, status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Warn("UpdateOrderStatus заказ отсутствует", slog.String("order_id", orderID), slog.String("status", status))
			return domain.ErrRowsNotFound
		}
		r.log.Error("UpdateOrderStatus ошибка обновления статуса", slog.String("order_id", orderID), slog.String("status", status), slog.Any("err", err))
		return err
	}

	r.log.Debug("UpdateOrderStatus завершено успешно", slog.String("order_id", orderID), slog.String("status", status))
	return nil
}

func (r *OrderRepoPostgres) GetOrder(ctx context.Context, orderID string) (*domain.OrderInfo, error) {
	r.log.Debug("GetOrder начало обработки", slog.String("order_id", orderID))

	rows, err := r.db.Query(ctx, getOrder, orderID)
	if err != nil {
		r.log.Error("GetOrder ошибка бд", slog.String("order_id", orderID), slog.Any("err", err))
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
			r.log.Error("GetOrder ошибка при декодировании данных", slog.String("order_id", orderID), slog.Any("err", err))
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetOrder ошибка после чтения строк", slog.String("order_id", orderID), slog.Any("err", err))
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn("GetOrder пустой ответ", slog.String("order_id", orderID))
		return nil, domain.ErrRowsNotFound
	}

	order.Items = items
	r.log.Debug("GetOrder завершено успешно", slog.String("order_id", orderID), slog.Int("items_count", len(items)))
	return &order, nil
}

func (r *OrderRepoPostgres) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	r.log.Debug("GetOrdersUser начало обработки", slog.String("user_id", userID))

	rows, err := r.db.Query(ctx, getUserOrders, userID)
	if err != nil {
		r.log.Error("GetOrdersUser ошибка бд", slog.String("user_id", userID), slog.Any("err", err))
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
			r.log.Error("GetOrdersUser ошибка при декодировании данных", slog.String("user_id", userID), slog.Any("err", err))
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetOrdersUser ошибка после чтения строк", slog.String("user_id", userID), slog.Any("err", err))
		return nil, err
	}

	if len(orders) == 0 {
		r.log.Debug("GetOrdersUser пустой ответ", slog.String("user_id", userID))
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetOrdersUser завершено успешно", slog.String("user_id", userID), slog.Int("orders_count", len(orders)))
	return orders, nil
}
