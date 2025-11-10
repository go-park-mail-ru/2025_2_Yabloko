package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
)

//go:embed sql/cart/get_items.sql
var getCartItems string

//go:embed sql/cart/delete_items.sql
var deleteCartItems string

//go:embed sql/cart/insert_item.sql
var insertCartItems string

type CartRepoPostgres struct {
	db  PgxIface
	log logger.Logger
}

func NewCartRepoPostgres(db PgxIface, log logger.Logger) *CartRepoPostgres {
	return &CartRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *CartRepoPostgres) GetCartItems(ctx context.Context, userID string) ([]*domain.CartItem, error) {
	r.log.Debug("🔍 GetCartItems", map[string]interface{}{"userID": userID})

	rows, err := r.db.Query(ctx, getCartItems, userID)
	if err != nil {
		r.log.Error("❌ GetCartItems: query failed", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItem
	for rows.Next() {
		var item domain.CartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.CardImg, &item.Price, &item.Quantity); err != nil {
			r.log.Error("❌ GetCartItems: scan failed", map[string]interface{}{
				"userID": userID,
				"err":    err,
			})
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("❌ GetCartItems: rows iteration error", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn("📭 Cart is empty", map[string]interface{}{"userID": userID})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("✅ GetCartItems succeeded", map[string]interface{}{
		"userID":    userID,
		"itemCount": len(items),
	})
	return items, nil
}

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, userID string) error {
	r.log.Debug("🧹 DeleteCartItems", map[string]interface{}{"userID": userID})

	_, err := r.db.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error("❌ DeleteCartItems failed", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return err
	}

	r.log.Debug("✅ DeleteCartItems succeeded", map[string]interface{}{"userID": userID})
	return nil
}

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error {
	r.log.Debug("🔄 UpdateCartItems", map[string]interface{}{
		"userID": userID,
		"itemCount": func() int {
			if newItems == nil {
				return 0
			}
			return len(newItems.Items)
		}(),
	})

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error("❌ Failed to begin transaction", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return err
	}
	defer func() {
		if tx != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr.Error() != "tx is not open" {
				r.log.Warn("⚠️ Rollback failed", map[string]interface{}{
					"userID": userID,
					"err":    rollbackErr,
				})
			}
		}
	}()

	// Шаг 1: Получаем или создаём корзину
	var cartID string
	err = tx.QueryRow(ctx, "SELECT id FROM cart WHERE user_id = $1", userID).Scan(&cartID)
	if err != nil {
		r.log.Debug("🆕 Creating new cart (not found)", map[string]interface{}{"userID": userID})
		cartID = uuid.New().String()
		_, err = tx.Exec(ctx, "INSERT INTO cart (id, user_id) VALUES ($1, $2)", cartID, userID)
		if err != nil {
			r.log.Error("❌ Failed to create cart", map[string]interface{}{
				"userID": userID,
				"err":    err,
			})
			return err
		}
	} else {
		r.log.Debug("📥 Using existing cart", map[string]interface{}{"cartID": cartID})
	}

	// Шаг 2: Удаляем старые элементы
	_, err = tx.Exec(ctx, "DELETE FROM cart_item WHERE cart_id = $1", cartID)
	if err != nil {
		r.log.Error("❌ Failed to delete old cart items", map[string]interface{}{
			"cartID": cartID,
			"err":    err,
		})
		return err
	}

	// Шаг 3: Вставляем новые элементы
	if newItems == nil || len(newItems.Items) == 0 {
		r.log.Warn("⚠️ Skipping item insertion: no items provided", map[string]interface{}{
			"cartID": cartID,
			"userID": userID,
		})
	} else {
		for i, item := range newItems.Items {
			if item.ID == "" {
				r.log.Error("🔥 Item ID is empty", map[string]interface{}{
					"index":  i,
					"cartID": cartID,
				})
				return fmt.Errorf("item[%d]: ID is empty", i)
			}

			_, err := tx.Exec(ctx, insertCartItems, uuid.New().String(), cartID, item.ID, item.Quantity)
			if err != nil {
				r.log.Error("❌ Failed to insert cart item", map[string]interface{}{
					"index":  i,
					"itemID": item.ID,
					"cartID": cartID,
					"err":    err,
				})
				return err
			}
		}
	}

	// Шаг 4: Коммит
	if err := tx.Commit(ctx); err != nil {
		r.log.Error("❌ Transaction commit failed", map[string]interface{}{
			"cartID": cartID,
			"userID": userID,
			"err":    err,
		})
		return err
	}

	r.log.Debug("✅ UpdateCartItems succeeded", map[string]interface{}{
		"cartID":    cartID,
		"userID":    userID,
		"itemCount": len(newItems.Items),
	})
	return nil
}
