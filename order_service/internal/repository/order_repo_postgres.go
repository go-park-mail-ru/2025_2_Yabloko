package repository

import (
	"apple_backend/order_service/internal/domain"
	"apple_backend/pkg/logger"
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

func (r *OrderRepoPostgres) CreateOrder(ctx context.Context, userID string, isFast bool, comment string) (string, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo CreateOrder start",
		slog.String("user_id", userID),
		slog.Bool("is_fast", isFast),
		slog.String("comment", comment),
	)

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

	orderID := uuid.New().String()
	_, err = tx.Exec(ctx, insertEmptyOrder, orderID, userID, isFast, comment)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder create order failed",
			slog.String("user_id", userID),
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	_, err = tx.Exec(ctx, insertItemOrder, orderID, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder transfer items failed",
			slog.String("user_id", userID),
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	_, err = tx.Exec(ctx, updateOrderTotal, orderID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder update total failed",
			slog.String("user_id", userID),
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	_, err = tx.Exec(ctx, deleteCartItemsForOrder, userID)
	if err != nil {
		log.ErrorContext(ctx, "repo CreateOrder clear cart failed",
			slog.String("user_id", userID),
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	if err = tx.Commit(ctx); err != nil {
		log.ErrorContext(ctx, "repo CreateOrder transaction commit failed",
			slog.String("user_id", userID),
			slog.String("order_id", orderID),
			slog.Any("err", err))
		return "", domain.ErrInternalServer
	}

	log.DebugContext(ctx, "repo CreateOrder success",
		slog.String("user_id", userID),
		slog.String("order_id", orderID))
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
	storeMap := make(map[string]*domain.StoreInfo)

	for rows.Next() {
		var (
			storeID      string
			storeName    string
			storeCardImg string
			item         domain.OrderItemInfo
			isFast       bool
			comment      *string
		)
		err = rows.Scan(
			&order.ID,
			&order.Total,
			&order.Status,
			&isFast,
			&comment,
			&order.CreatedAt,
			&storeID,
			&storeName,
			&storeCardImg,
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

		order.IsFast = isFast
		if comment != nil {
			order.Comment = *comment
		}

		if _, exists := storeMap[storeID]; !exists {
			storeMap[storeID] = &domain.StoreInfo{
				ID:      storeID,
				Name:    storeName,
				CardImg: storeCardImg,
				Items:   make([]*domain.OrderItemInfo, 0),
			}
		}
		storeMap[storeID].Items = append(storeMap[storeID].Items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "repo GetOrder rows error", slog.String("order_id", orderID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	if len(storeMap) == 0 {
		log.WarnContext(ctx, "repo GetOrder no items found", slog.String("order_id", orderID))
		return nil, domain.ErrRowsNotFound
	}

	order.Stores = make([]*domain.StoreInfo, 0, len(storeMap))
	for _, store := range storeMap {
		order.Stores = append(order.Stores, store)
	}

	log.DebugContext(ctx, "repo GetOrder success", slog.String("order_id", orderID), slog.Int("stores_count", len(order.Stores)))
	return &order, nil
}

func (r *OrderRepoPostgres) GetOrdersUser(ctx context.Context, filter *domain.OrderFilter) ([]*domain.Order, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo GetOrdersUser params",
		slog.String("user_id", filter.UserID),
		slog.Int("limit", filter.Limit),
		slog.String("last_id", filter.LastID))

	var lastID interface{}
	if filter.LastID != "" {
		lastID = filter.LastID
	} else {
		lastID = nil
	}

	rows, err := r.db.Query(ctx, getUserOrders, filter.UserID, lastID, filter.Limit)
	if err != nil {
		log.ErrorContext(ctx, "repo GetOrdersUser query failed",
			slog.String("user_id", filter.UserID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var comment *string
		err = rows.Scan(
			&order.ID,
			&order.Status,
			&order.Total,
			&order.CreatedAt,
			&order.IsFast,
			&comment,
			&order.StoreID,
			&order.StoreName,
		)
		if err != nil {
			log.ErrorContext(ctx, "repo GetOrdersUser scan failed",
				slog.String("user_id", filter.UserID), slog.Any("err", err))
			return nil, domain.ErrInternalServer
		}
		if comment != nil {
			order.Comment = *comment
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "repo GetOrdersUser rows error",
			slog.String("user_id", filter.UserID), slog.Any("err", err))
		return nil, domain.ErrInternalServer
	}

	if len(orders) == 0 {
		log.DebugContext(ctx, "repo GetOrdersUser no orders found", slog.String("user_id", filter.UserID))
		return []*domain.Order{}, nil
	}

	log.DebugContext(ctx, "repo GetOrdersUser success",
		slog.String("user_id", filter.UserID), slog.Int("orders_count", len(orders)))
	return orders, nil
}
