package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

//go:embed sql/cart/get_items.sql
var getCartItems string

//go:embed sql/cart/delete_items.sql
var deleteCartItems string

//go:embed sql/cart/insert_item.sql
var insertCartItems string

type CartRepoPostgres struct {
	db PgxIface
}

func NewCartRepoPostgres(db PgxIface) *CartRepoPostgres {
	return &CartRepoPostgres{
		db: db,
	}
}

func (r *CartRepoPostgres) GetCartItems(ctx context.Context, userID string) ([]*domain.CartItem, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetCartItems начало обработки", slog.String("user_id", userID))

	rows, err := r.db.Query(ctx, getCartItems, userID)
	if err != nil {
		log.ErrorContext(ctx, "GetCartItems query failed",
			slog.Any("err", err),
			slog.String("user_id", userID))
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItem
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.CardImg, &item.Price, &item.Quantity); err != nil {
			log.ErrorContext(ctx, "GetCartItems scan failed",
				slog.Any("err", err),
				slog.String("user_id", userID))
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetCartItems rows iteration error",
			slog.Any("err", err),
			slog.String("user_id", userID))
		return nil, err
	}

	if len(items) == 0 {
		log.DebugContext(ctx, "GetCartItems cart is empty", slog.String("user_id", userID))
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetCartItems завершено успешно",
		slog.String("user_id", userID),
		slog.Int("items_count", len(items)))
	return items, nil
}

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, userID string) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "DeleteCartItems начало обработки", slog.String("user_id", userID))

	_, err := r.db.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		log.ErrorContext(ctx, "DeleteCartItems failed",
			slog.Any("err", err),
			slog.String("user_id", userID))
		return err
	}

	log.DebugContext(ctx, "DeleteCartItems завершено успешно", slog.String("user_id", userID))
	return nil
}

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error {
	log := logger.FromContext(ctx)

	itemsCount := 0
	if newItems != nil {
		itemsCount = len(newItems.Items)
	}

	log.DebugContext(ctx, "UpdateCartItems начало обработки",
		slog.String("user_id", userID),
		slog.Int("items_count", itemsCount))

	if newItems != nil && len(newItems.Items) > 0 {
		for i, item := range newItems.Items {
			var exists bool
			checkQuery := `SELECT EXISTS(SELECT 1 FROM store_item WHERE id = $1)`
			err := r.db.QueryRow(ctx, checkQuery, item.ID).Scan(&exists)
			if err != nil {
				log.ErrorContext(ctx, "UpdateCartItems ошибка проверки store_item",
					slog.Any("err", err),
					slog.String("item_id", item.ID))
				return err
			}
			if !exists {
				log.WarnContext(ctx, "UpdateCartItems store_item не найден",
					slog.String("item_id", item.ID),
					slog.Int("item_index", i))
				return domain.ErrRowsNotFound
			}
		}
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.ErrorContext(ctx, "UpdateCartItems transaction begin failed",
			slog.Any("err", err),
			slog.String("user_id", userID))
		return err
	}
	defer func() {
		if tx != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr.Error() != "tx is not open" {
				log.WarnContext(ctx, "UpdateCartItems rollback failed",
					slog.Any("err", rollbackErr),
					slog.String("user_id", userID))
			}
		}
	}()

	var cartID string
	err = tx.QueryRow(ctx, "SELECT id FROM cart WHERE user_id = $1", userID).Scan(&cartID)
	if err != nil {
		log.DebugContext(ctx, "UpdateCartItems creating new cart", slog.String("user_id", userID))
		cartID = uuid.New().String()
		_, err = tx.Exec(ctx, "INSERT INTO cart (id, user_id) VALUES ($1, $2)", cartID, userID)
		if err != nil {
			log.ErrorContext(ctx, "UpdateCartItems create cart failed",
				slog.Any("err", err),
				slog.String("user_id", userID))
			return err
		}
	} else {
		log.DebugContext(ctx, "UpdateCartItems using existing cart",
			slog.String("cart_id", cartID),
			slog.String("user_id", userID))
	}

	_, err = tx.Exec(ctx, "DELETE FROM cart_item WHERE cart_id = $1", cartID)
	if err != nil {
		log.ErrorContext(ctx, "UpdateCartItems delete old items failed",
			slog.Any("err", err),
			slog.String("cart_id", cartID),
			slog.String("user_id", userID))
		return err
	}

	if newItems == nil || len(newItems.Items) == 0 {
		log.DebugContext(ctx, "UpdateCartItems no items to insert",
			slog.String("cart_id", cartID),
			slog.String("user_id", userID))
	} else {
		for i, item := range newItems.Items {
			if item.ID == "" {
				log.ErrorContext(ctx, "UpdateCartItems empty item ID",
					slog.Int("item_index", i),
					slog.String("cart_id", cartID))
				return fmt.Errorf("item[%d]: ID is empty", i)
			}

			_, err := tx.Exec(ctx, insertCartItems, uuid.New().String(), cartID, item.ID, item.Quantity)
			if err != nil {
				log.ErrorContext(ctx, "UpdateCartItems insert item failed",
					slog.Any("err", err),
					slog.Int("item_index", i),
					slog.String("item_id", item.ID),
					slog.String("cart_id", cartID),
					slog.Int("quantity", item.Quantity))
				return err
			}
		}
		log.DebugContext(ctx, "UpdateCartItems items inserted successfully",
			slog.String("cart_id", cartID),
			slog.Int("items_count", len(newItems.Items)))
	}

	if err := tx.Commit(ctx); err != nil {
		log.ErrorContext(ctx, "UpdateCartItems transaction commit failed",
			slog.Any("err", err),
			slog.String("cart_id", cartID),
			slog.String("user_id", userID))
		return err
	}

	log.DebugContext(ctx, "UpdateCartItems завершено успешно",
		slog.String("cart_id", cartID),
		slog.String("user_id", userID),
		slog.Int("items_count", itemsCount))
	return nil
}
