package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"

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
	r.log.Debug("GetCartItems начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getCartItems, userID)
	if err != nil {
		r.log.Error("GetCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItem
	for rows.Next() {
		var item domain.CartItem
		err = rows.Scan(&item.ID, &item.Name, &item.CardImg, &item.Price, &item.Quantity)
		if err != nil {
			r.log.Error("GetCartItems ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetCartItems ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn("GetCartItems пустой ответ", map[string]interface{}{"id": userID})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetCartItems завершено успешно", map[string]interface{}{})
	return items, nil
}

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, userID string) error {
	r.log.Debug("DeleteCartItems начало обработки", map[string]interface{}{})

	_, err := r.db.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error("DeleteCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return err
	}

	r.log.Debug("DeleteCartItems завершено успешно", map[string]interface{}{})
	return nil
}

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error {
	r.log.Debug("UpdateCartItems начало обработки", map[string]interface{}{})

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Удаляем текущие товары пользователя
	_, err = tx.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error("UpdateCartItems ошибка удаления записи", map[string]interface{}{"err": err, "id": userID})
		return err
	}

	// 2. Вставляем новые
	for _, item := range newItems.Items {
		_, err = tx.Exec(ctx, insertCartItems, uuid.New().String(), userID, item.ID, item.Quantity)
		if err != nil {
			r.log.Error("UpdateCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error("UpdateCartItems ошибка закрытия транзакции", map[string]interface{}{"err": err, "id": userID})
		return domain.ErrInternalServer
	}

	r.log.Debug("UpdateCartItems завершено успешно", map[string]interface{}{})
	return nil
}
