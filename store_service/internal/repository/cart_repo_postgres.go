package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"

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

func (r *CartRepoPostgres) GetCartItems(ctx context.Context, userID string) (string, []*domain.CartItem, error) {
	r.log.Debug(ctx, "GetCartItems начало обработки", map[string]interface{}{})
	query := `
		select
		    c.id as cart_id,
		    si.id as id,
			it.name as name,
			it.card_img as card_img,
			si.price as price,
			ci.quantity as quantity
		from cart c
		join cart_item ci on ci.cart_id = c.id
		join store_item si on si.id = ci.store_item_id
		join item it on it.id = si.item_id
		WHERE c.user_id = $1
		ORDER BY ci.created_at;
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.log.Error(ctx, "GetCartItems ошибка бд", map[string]interface{}{"err": err, "id": userID})
		return "", nil, err
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
			return "", nil, err
		}

		if cartID == "" {
			cartID = rowCartID
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetCartItems ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": userID})
		return "", nil, err
	}

	if len(items) == 0 {
		r.log.Warn(ctx, "GetCartItems пустой ответ", map[string]interface{}{"id": userID})
		return "", nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetCartItems завершено успешно", map[string]interface{}{})
	return cartID, items, nil
}

func (r *CartRepoPostgres) DeleteCartItems(ctx context.Context, id string) error {
	r.log.Debug(ctx, "DeleteCartItems начало обработки", map[string]interface{}{})
	query := `
		delete from cart_item
		where cart_id = $1;
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "DeleteCartItems ошибка бд", map[string]interface{}{"err": err, "id": id})
		return err
	}

	r.log.Debug(ctx, "DeleteCartItems завершено успешно", map[string]interface{}{})
	return nil
}

func (r *CartRepoPostgres) UpdateCartItems(ctx context.Context, id string, newItems *domain.CartUpdate) error {
	r.log.Debug(ctx, "UpdateCartItems начало обработки", map[string]interface{}{})
	// выполняем два действия
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		delete from cart_item
		where cart_id = $1;
	`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "UpdateCartItems ошибка удаления записи", map[string]interface{}{"err": err, "id": id})
		return err
	}

	query = `
		insert into cart_item (id, cart_id, store_item_id, quantity)
		values ($1, $2, $3, $4);
	`
	for _, item := range newItems.Items {
		_, err := tx.Exec(ctx, query, uuid.New().String(), id, item.ID, item.Quantity)
		if err != nil {
			r.log.Error(ctx, "UpdateCartItems ошибка бд", map[string]interface{}{"err": err, "id": id})
			return err
		}
	}

	tx.Commit(ctx)
	r.log.Debug(ctx, "UpdateCartItems завершено успешно", map[string]interface{}{})
	return nil
}
