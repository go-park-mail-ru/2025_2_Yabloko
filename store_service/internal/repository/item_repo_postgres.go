package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
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

//go:embed sql/item/get_types.sql
var getItemTypes string

func (r *ItemRepoPostgres) GetItemTypes(ctx context.Context, id string) ([]*domain.ItemType, error) {
	r.log.Debug(ctx, "GetItemTypes начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getItemTypes, id)
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
		r.log.Warn(ctx, "GetItemTypes пустой ответ", map[string]interface{}{"id": id})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug(ctx, "GetItemTypes завершено успешно", map[string]interface{}{})
	return itemTypes, nil
}

//go:embed sql/item/get_items.sql
var getItems string

func (r *ItemRepoPostgres) GetItems(ctx context.Context, id string) ([]*domain.Item, error) {
	r.log.Debug(ctx, "GetItems начало обработки", map[string]interface{}{})

	// если у товара несколько типов, то будет несколько строк с этим товаром
	rows, err := r.db.Query(ctx, getItems, id)
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
