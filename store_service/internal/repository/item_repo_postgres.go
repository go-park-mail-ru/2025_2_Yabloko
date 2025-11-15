package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"log/slog"
)

//go:embed sql/item/get_types.sql
var getItemTypes string

//go:embed sql/item/get_items.sql
var getItems string

type ItemRepoPostgres struct {
	db PgxIface
}

func NewItemRepoPostgres(db PgxIface) *ItemRepoPostgres {
	return &ItemRepoPostgres{
		db: db,
	}
}

func (r *ItemRepoPostgres) GetItemTypes(ctx context.Context, storeID string) ([]*domain.ItemType, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetItemTypes начало обработки", slog.String("store_id", storeID))

	rows, err := r.db.Query(ctx, getItemTypes, storeID)
	if err != nil {
		log.ErrorContext(ctx, "GetItemTypes ошибка бд", slog.Any("err", err), slog.String("store_id", storeID))
		return nil, err
	}
	defer rows.Close()

	var itemTypes []*domain.ItemType
	for rows.Next() {
		var itemType domain.ItemType
		err = rows.Scan(&itemType.ID, &itemType.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetItemTypes ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		itemTypes = append(itemTypes, &itemType)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetItemTypes ошибка после чтения строк", slog.Any("err", err), slog.String("store_id", storeID))
		return nil, err
	}

	if len(itemTypes) == 0 {
		log.DebugContext(ctx, "GetItemTypes пустой ответ", slog.String("store_id", storeID))
		return []*domain.ItemType{}, nil // возвращаем пустой массив вместо ошибки
	}

	log.DebugContext(ctx, "GetItemTypes завершено успешно",
		slog.String("store_id", storeID),
		slog.Int("types_count", len(itemTypes)))
	return itemTypes, nil
}

func (r *ItemRepoPostgres) GetItems(ctx context.Context, itemTypeID string) ([]*domain.Item, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetItems начало обработки", slog.String("type_id", itemTypeID))

	rows, err := r.db.Query(ctx, getItems, itemTypeID)
	if err != nil {
		log.ErrorContext(ctx, "GetItems ошибка бд", slog.Any("err", err), slog.String("type_id", itemTypeID))
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
			log.ErrorContext(ctx, "GetItems ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetItems ошибка после чтения строк", slog.Any("err", err), slog.String("type_id", itemTypeID))
		return nil, err
	}

	if len(items) == 0 {
		log.DebugContext(ctx, "GetItems пустой ответ", slog.String("type_id", itemTypeID))
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetItems завершено успешно",
		slog.String("type_id", itemTypeID),
		slog.Int("items_count", len(items)))
	return items, nil
}
