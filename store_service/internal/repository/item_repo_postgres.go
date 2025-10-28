package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
)

type ItemRepoPostgres struct {
	db  PgxIface
	log *logger.Logger
}

func NewItemRepoPostgres(db PgxIface, log *logger.Logger) *ItemRepoPostgres {
	return &ItemRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *ItemRepoPostgres) GetItemTypes(ctx context.Context, id string) ([]*domain.ItemType, error) {
	r.log.Debug(ctx, "GetItemTypes начало обработки", map[string]interface{}{})
	query := `
		select type.id, type.name
		from store_item join item_type on store_item.item_id = item_type.item_id
		join type on store_item.type_id = type.id
		where store_item.store_id = $1
	`

	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "GetItemTypes ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var itemTypes []*domain.ItemType
	for rows.Next() {
		var itemType domain.ItemType
		err = rows.Scan(&itemType.ID, &itemType.Name)
		if err != nil {
			r.log.Error(ctx, "GetItemTypes ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		itemTypes = append(itemTypes, &itemType)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetItemTypes ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(itemTypes) == 0 {
		r.log.Debug(ctx, "GetItemTypes пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetItemTypes завершено успешно", map[string]interface{}{})
	return itemTypes, nil
}

func (r *ItemRepoPostgres) GetItems(ctx context.Context, id string) ([]*domain.Item, error) {
	r.log.Debug(ctx, "GetItems начало обработки", map[string]interface{}{})
	query := `
		select store_item.id, item.name, store_item.price, item.description, item.card_img, item_type.type_id
		from store_item join item on store_item.item_id = item.id
		join item_type on item.id = item_type.item_id
		where store_item.store_id = $1
	`

	// если у товара несколько типов, то будет несколько строк с этим товаром
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		r.log.Error(ctx, "GetItems ошибка бд", map[string]interface{}{"err": err, "id": id})
		return nil, err
	}
	defer rows.Close()

	var items []*domain.Item
	for rows.Next() {
		item := &domain.Item{}
		err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Description,
			&item.CardImg,
			&item.TypeID,
		)
		if err != nil {
			r.log.Error(ctx, "GetItems ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "GetItems ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": id})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Debug(ctx, "GetItems пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetItems завершено успешно", map[string]interface{}{})
	return items, nil
}
