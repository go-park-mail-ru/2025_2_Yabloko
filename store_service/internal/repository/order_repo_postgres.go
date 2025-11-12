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

//go:embed sql/cart/delete_items.sql
var deleteCartItemsForOrder string

type OrderRepoPostgres struct {
	db PgxIface
}

func NewOrderRepoPostgres(db PgxIface) *OrderRepoPostgres {
	return &OrderRepoPostgres{
		db: db,
	}
}

func (r *OrderRepoPostgres) GetOrderUserID(ctx context.Context, orderID string) (string, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetOrderUserID start", slog.String("order_id", orderID))

	var userID string
	err := r.db.QueryRow(ctx, getOrderUser, orderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo GetOrderUserID order not found", slog.String("order_id", orderID))
			return "", domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo GetOrderUserID query failed", slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo GetOrderUserID success", slog.String("order_id", orderID))
	return userID, nil
}

func (r *OrderRepoPostgres) CreateOrder(ctx context.Context, userID string) (string, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo CreateOrder start", slog.String("user_id", userID))

	// проверка есть ли товары в корзине
	var cnt int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM cart_item ci JOIN cart c ON c.id = ci.cart_id WHERE c.user_id = $1", userID).Scan(&cnt)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder cart check failed", slog.String("user_id", userID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}
	if cnt == 0 {
		log.WarnContext(ctx, "repo CreateOrder cart is empty", slog.String("user_id", userID))
		return "", domain.ErrCartEmpty
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder transaction begin failed", slog.String("user_id", userID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}
	defer tx.Rollback(ctx)

	// 1 - создаем заказ
	orderID := uuid.New().String()
	_, err = tx.Exec(ctx, insertEmptyOrder, orderID, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder create order failed", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	// 2 - переносим товары из корзины
	_, err = tx.Exec(ctx, insertItemOrder, orderID, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder transfer items failed", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	// 3 - обновляем итоговую сумму
	_, err = tx.Exec(ctx, updateOrderTotal, orderID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder update total failed", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	// 4 - очищаем корзину
	_, err = tx.Exec(ctx, deleteCartItemsForOrder, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder clear cart failed", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder transaction commit failed", slog.String("user_id", userID), slog.String("order_id", orderID), slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo CreateOrder success", slog.String("user_id", userID), slog.String("order_id", orderID))
	return orderID, nil
}

func (r *OrderRepoPostgres) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo UpdateOrderStatus start", slog.String("order_id", orderID), slog.String("status", status))

	_, err := r.db.Exec(ctx, updateOrderStatus, orderID, status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.WarnContext(ctx, "repo UpdateOrderStatus order not found", slog.String("order_id", orderID), slog.String("status", status))
			return domain.ErrRowsNotFound
		}
		log.ErrorContext(ctx, "repo UpdateOrderStatus update failed", slog.String("order_id", orderID), slog.String("status", status), slog.Any("err", err))
		return domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo UpdateOrderStatus success", slog.String("order_id", orderID), slog.String("status", status))
	return nil
}

func (r *OrderRepoPostgres) GetOrder(ctx context.Context, orderID string) (*domain.OrderInfo, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetOrder start", slog.String("order_id", orderID))

	rows, err := r.db.Query(ctx, getOrder, orderID)
	if err != nil {
		log.ErrorContext(ctx, "repo GetOrder query failed", slog.String("order_id", orderID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
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
			log.ErrorContext(ctx, "repo GetOrder scan failed", slog.String("order_id", orderID), slog.Any("err", err))
			return nil, domain.ErrInternalServer
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "repo GetOrder rows error", slog.String("order_id", orderID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	if len(items) == 0 {
		log.WarnContext(ctx, "repo GetOrder no items found", slog.String("order_id", orderID))
		return nil, domain.ErrRowsNotFound
	}

	order.Items = items
	log.DebugContext(ctx, "repo GetOrder success", slog.String("order_id", orderID), slog.Int("items_count", len(items)))
	return &order, nil
}

func (r *OrderRepoPostgres) GetOrdersUser(ctx context.Context, userID string) ([]*domain.Order, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetOrdersUser start", slog.String("user_id", userID))

	rows, err := r.db.Query(ctx, getUserOrders, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo GetOrdersUser query failed", slog.String("user_id", userID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
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
			log.ErrorContext(ctx, "repo GetOrdersUser scan failed", slog.String("user_id", userID), slog.Any("err", err))
			return nil, domain.ErrInternalServer
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "repo GetOrdersUser rows error", slog.String("user_id", userID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	if len(orders) == 0 {
		log.DebugContext(ctx, "repo GetOrdersUser no orders found", slog.String("user_id", userID))
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "repo GetOrdersUser success", slog.String("user_id", userID), slog.Int("orders_count", len(orders)))
	return orders, nil
}
