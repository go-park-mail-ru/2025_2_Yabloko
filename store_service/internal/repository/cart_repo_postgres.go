package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"

	"github.com/google/uuid"
)

type CartRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewCartRepoPostgres(db PgxIface, log *logger.Logger) *CartRepoPostgres {
	return &CartRepoPostgres{
		db:  db,
		log: log,
	}
}

//go:embed sql/cart/get_items.sql
var getCartItems string

func (r *CartRepoPostgres) GetCartItems(ctx context.Context, userID string) ([]*domain.CartItem, error) {
	r.log.Debug(ctx, "GetCartItems начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getCartItems, userID)
	if err != nil {
		r.log.Error(ctx, "GetCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItem
	var cartID string
	for rows.Next() {
		var item domain.CartItem
		var rowCartID string
		err = rows.Scan(&rowCartID, &item.ID, &item.Name, &item.CardImg, &item.Price, &item.Quantity)
		if err != nil {
			r.log.Error(ctx, "GetCartItems ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}

		if cartID == "" {
			cartID = rowCartID
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetCartItems ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": userID})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Warn(ctx, "GetCartItems пустой ответ", map[string]interface{}{"id": userID})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetCartItems завершено успешно", map[string]interface{}{})
	return items, nil
}

//go:embed sql/cart/delete_items.sql
var deleteCartItems string

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, userID string) error {
	r.log.Debug(ctx, "DeleteCartItems начало обработки", map[string]interface{}{})

	_, err := r.db.Exec(ctx, deleteCartItems, userID)
	if err != nil {
		r.log.Error(ctx, "DeleteCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return err
	}

	r.log.Debug(ctx, "DeleteCartItems завершено успешно", map[string]interface{}{})
	return nil
}

//go:embed sql/cart/insert_item.sql
var insertCartItems string

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, userID string, newItems *domain.CartUpdate) error {
	r.log.Debug(ctx, "UpdateCartItems начало обработки", map[string]interface{}{})
	// выполняем два действия
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var cartID string
	err = tx.QueryRow(ctx, deleteCartItems, userID).Scan(&cartID)
	if err != nil {
		r.log.Error(ctx, "UpdateCartItems ошибка удаления записи", map[string]interface{}{"err": err, "id": userID})
		return err
	}

	for _, item := range newItems.Items {
		_, err = tx.Exec(ctx, insertCartItems, uuid.New().String(), cartID, item.ID, item.Quantity)
		if err != nil {
			r.log.Error(ctx, "UpdateCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.log.Error(ctx, "UpdateCartItems ошибка закрытия транзакции", map[string]interface{}{"err": err, "id": userID})
		return domain.ErrInternalServer
	}

	r.log.Debug(ctx, "UpdateCartItems завершено успешно", map[string]interface{}{})
	return nil
}
