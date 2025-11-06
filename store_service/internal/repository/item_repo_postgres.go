package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
)

//go:embed sql/item/get_types.sql
var getItemTypes string

//go:embed sql/item/get_items.sql
var getItems string

type ItemRepoPostgres struct {
	db  PgxIface
	log logger.Logger
}

func NewItemRepoPostgres(db PgxIface, log logger.Logger) *ItemRepoPostgres {
	return &ItemRepoPostgres{
		db:  db,
		log: log,
	}
}

func (r *ItemRepoPostgres) GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error) {
	r.log.Debug("GetItemTypes начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getItemTypes, storeID)
	if err != nil {
		r.log.Error("GetItemTypes ошибка бд", map[string]interface{}{"err": err, "id": storeID})
		return nil, err
	}
	defer rows.Close()

	var itemTypes []*domain.ItemType
	for rows.Next() {
		var itemType domain.ItemType
		err = rows.Scan(&itemType.ID, &itemType.Name)
		if err != nil {
			r.log.Error("GetItemTypes ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		itemTypes = append(itemTypes, &itemType)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetItemTypes ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": storeID})
		return nil, err
	}

	if len(itemTypes) == 0 {
		r.log.Warn("GetItemTypes пустой ответ", map[string]interface{}{"id": storeID})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetItemTypes завершено успешно", map[string]interface{}{})
	return itemTypes, nil
}

func (r *ItemRepoPostgres) GetItems(ctx context.Context, itemTypeID string) ([]*domain.Item, error) {
	r.log.Debug("GetItems начало обработки", map[string]interface{}{})

	rows, err := r.db.Query(ctx, getItems, itemTypeID)
	if err != nil {
		r.log.Error("GetItems ошибка бд", map[string]interface{}{"err": err, "id": itemTypeID})
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
			r.log.Error("GetItems ошибка при декодировании данных",
				map[string]interface{}{"err": err, "rows": rows})
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("GetItems ошибка после чтения строк",
			map[string]interface{}{"err": err, "id": itemTypeID})
		return nil, err
	}

	if len(items) == 0 {
		r.log.Debug("GetItems пустой ответ", map[string]interface{}{"id": itemTypeID})
		return nil, domain.ErrRowsNotFound
	}

	r.log.Debug("GetItems завершено успешно", map[string]interface{}{})
	return items, nil
}
