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
	r.log.Debug("🔍 GetCartItems начало обработки", map[string]interface{}{
		"userID": userID,
	})

	rows, err := r.db.Query(ctx, getCartItems, userID)
	if err != nil {
		r.log.Error("❌ GetCartItems: ошибка выполнения запроса", map[string]interface{}{
			"userID": userID,
			"query":  getCartItems,
			"err":    err,
		})
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItem
	for rows.Next() {
		var item domain.CartItem
		scanErr := rows.Scan(&item.ID, &item.Name, &item.CardImg, &item.Price, &item.Quantity)
		if scanErr != nil {
			r.log.Error("❌ GetCartItems: ошибка при Scan строки", map[string]interface{}{
				"userID":  userID,
				"scanErr": scanErr,
				"itemRaw": fmt.Sprintf("%+v", item),
			})
			return nil, scanErr
		}
		r.log.Debug("📥 Получен элемент корзины", map[string]interface{}{
			"userID":   userID,
			"itemID":   item.ID,
			"name":     item.Name,
			"price":    item.Price,
			"quantity": item.Quantity,
		})
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("❌ GetCartItems: ошибка после итерации по rows", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn("📭 GetCartItems: пустой результат (корзина пуста)", map[string]interface{}{
			"userID": userID,
		})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("✅ GetCartItems завершён успешно", map[string]interface{}{
		"userID":    userID,
		"itemCount": len(items),
		"items": func() []map[string]interface{} {
			var out []map[string]interface{}
			for _, it := range items {
				out = append(out, map[string]interface{}{
					"id":       it.ID,
					"name":     it.Name,
					"price":    it.Price,
					"quantity": it.Quantity,
				})
			}
			return out
		}(),
	})
	return items, nil
}

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, userID string) error {
	r.log.Debug("🧹 DeleteCartItems начало обработки", map[string]interface{}{
		"userID": userID,
	})

	_, err := r.db.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error("❌ DeleteCartItems: ошибка выполнения DELETE", map[string]interface{}{
			"userID": userID,
			"query":  deleteCartItems,
			"err":    err,
		})
		return err
	}

	r.log.Debug("✅ DeleteCartItems завершено успешно", map[string]interface{}{
		"userID": userID,
	})
	return nil
}

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error {
	r.log.Debug("🔄 UpdateCartItems начало обработки", map[string]interface{}{
		"userID": userID,
		"newItems": func() map[string]interface{} {
			if newItems == nil {
				return map[string]interface{}{"items": []string{}}
			}
			itemsLog := make([]map[string]interface{}, len(newItems.Items))
			for i, item := range newItems.Items {
				itemsLog[i] = map[string]interface{}{
					"index":    i,
					"itemID":   item.ID, // 🔴 Вот здесь будет видно, пустой ли ID
					"quantity": item.Quantity,
				}
			}
			return map[string]interface{}{
				"itemCount": len(newItems.Items),
				"items":     itemsLog,
			}
		}(),
	})

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error("❌ UpdateCartItems: не удалось начать транзакцию", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		return err
	}
	defer func() {
		if tx != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil && rollbackErr.Error() != "tx is not open" {
				r.log.Warn("⚠️ UpdateCartItems: ошибка при откате транзакции", map[string]interface{}{
					"userID":      userID,
					"rollbackErr": rollbackErr,
				})
			}
		}
	}()

	// 🔍 Шаг 1: Получаем cart_id по user_id
	r.log.Debug("🔍 UpdateCartItems: получение cart_id по user_id", map[string]interface{}{
		"userID": userID,
	})
	var cartID string
	err = tx.QueryRow(ctx, "SELECT id FROM cart WHERE user_id = $1", userID).Scan(&cartID)
	if err != nil {
		r.log.Warn("🟡 UpdateCartItems: корзина не найдена — создаём новую", map[string]interface{}{
			"userID": userID,
			"err":    err,
		})
		cartID = uuid.New().String()
		r.log.Debug("🆕 Создаём новую корзину", map[string]interface{}{
			"newCartID": cartID,
			"userID":    userID,
		})
		_, insertErr := tx.Exec(ctx, "INSERT INTO cart (id, user_id) VALUES ($1, $2)", cartID, userID)
		if insertErr != nil {
			r.log.Error("❌ UpdateCartItems: ошибка при создании корзины", map[string]interface{}{
				"userID":    userID,
				"cartID":    cartID,
				"insertErr": insertErr,
			})
			return insertErr
		}
		r.log.Debug("✅ Корзина создана успешно", map[string]interface{}{
			"cartID": cartID,
			"userID": userID,
		})
	} else {
		r.log.Debug("✅ Найдена существующая корзина", map[string]interface{}{
			"cartID": cartID,
			"userID": userID,
		})
	}

	// 🗑️ Шаг 2: Удаляем текущие элементы
	r.log.Debug("🧹 Удаляем старые элементы корзины", map[string]interface{}{
		"cartID": cartID,
	})
	_, delErr := tx.Exec(ctx, "DELETE FROM cart_item WHERE cart_id = $1", cartID)
	if delErr != nil {
		r.log.Error("❌ UpdateCartItems: ошибка при удалении старых элементов", map[string]interface{}{
			"cartID": cartID,
			"err":    delErr,
		})
		return delErr
	}
	r.log.Debug("✅ Старые элементы корзины удалены", map[string]interface{}{
		"cartID": cartID,
	})

	// ➕ Шаг 3: Вставляем новые элементы
	if newItems == nil || len(newItems.Items) == 0 {
		r.log.Warn("⚠️ UpdateCartItems: newItems пуст — завершаем без вставки", map[string]interface{}{
			"cartID": cartID,
			"userID": userID,
		})
	} else {
		r.log.Debug("➕ Вставляем новые элементы", map[string]interface{}{
			"cartID":    cartID,
			"itemCount": len(newItems.Items),
		})

		for i, item := range newItems.Items {
			r.log.Debug("📥 Вставка элемента", map[string]interface{}{
				"index":    i,
				"cartID":   cartID,
				"itemID":   item.ID, // 🔴 КРИТИЧЕСКИ ВАЖНО: проверьте это значение в логах!
				"quantity": item.Quantity,
				"sql":      insertCartItems,
				"params":   []interface{}{uuid.New().String(), cartID, item.ID, item.Quantity},
			})

			if item.ID == "" {
				r.log.Error("🔥 КРИТИЧЕСКАЯ ОШИБКА: item.ID пустой! Пропуск вставки.", map[string]interface{}{
					"index":    i,
					"cartID":   cartID,
					"quantity": item.Quantity,
				})
				// Можно вернуть ошибку явно:
				return fmt.Errorf("item[%d]: ID is empty", i)
			}

			_, insertErr := tx.Exec(ctx, insertCartItems, uuid.New().String(), cartID, item.ID, item.Quantity)
			if insertErr != nil {
				r.log.Error("❌ UpdateCartItems: ошибка при вставке элемента", map[string]interface{}{
					"index":       i,
					"cartID":      cartID,
					"itemID":      item.ID,
					"quantity":    item.Quantity,
					"insertErr":   insertErr,
					"errorDetail": fmt.Sprintf("%+v", insertErr),
				})
				return insertErr
			}
			r.log.Debug("✅ Элемент вставлен", map[string]interface{}{
				"index":    i,
				"itemID":   item.ID,
				"quantity": item.Quantity,
			})
		}
	}

	// 💾 Шаг 4: Коммит транзакции
	r.log.Debug("💾 Попытка коммита транзакции", map[string]interface{}{
		"cartID": cartID,
		"userID": userID,
	})
	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		r.log.Error("❌ UpdateCartItems: ошибка при коммите транзакции", map[string]interface{}{
			"cartID":    cartID,
			"userID":    userID,
			"commitErr": commitErr,
		})
		return commitErr
	}
	r.log.Debug("✅ UpdateCartItems завершён успешно", map[string]interface{}{
		"cartID":    cartID,
		"userID":    userID,
		"itemCount": len(newItems.Items),
	})

	return nil
}
