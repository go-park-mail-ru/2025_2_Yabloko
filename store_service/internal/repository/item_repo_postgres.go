package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

//go:embed sql/item/get_types.sql
var getItemTypes string

//go:embed sql/item/get_items.sql
var baseGetItems string

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
		return []*domain.ItemType{}, nil
	}

	log.DebugContext(ctx, "GetItemTypes завершено успешно",
		slog.String("store_id", storeID),
		slog.Int("types_count", len(itemTypes)))
	return itemTypes, nil
}

func generateGetItemsQuery(filter *domain.ItemFilter) (string, []any) {
	query := baseGetItems

	orderClauses := []string{}
	if filter.Sorted != "" {
		dir := "ASC"
		if filter.Desc {
			dir = "DESC"
		}
		switch filter.Sorted {
		case "name":
			orderClauses = append(orderClauses, fmt.Sprintf("i.name %s", dir))
		case "price":
			orderClauses = append(orderClauses, fmt.Sprintf("si.price %s", dir))
		}
	}
	if len(orderClauses) == 0 {
		orderClauses = append(orderClauses, "i.name ASC")
	}
	orderClauses = append(orderClauses, "si.id")

	query = query + "\nORDER BY " + strings.Join(orderClauses, ", ")

	args := []any{filter.StoreID}

	if len(filter.ItemTypes) == 0 {
		args = append(args, []string{})
	} else {
		args = append(args, filter.ItemTypes)
	}

	return query, args
}

func (r *ItemRepoPostgres) GetItems(ctx context.Context, filter *domain.ItemFilter) ([]*domain.ItemAgg, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetItems начало обработки",
		slog.String("store_id", filter.StoreID),
		slog.Any("item_types", filter.ItemTypes),
		slog.String("sorted", filter.Sorted),
		slog.Bool("desc", filter.Desc),
	)

	query, args := generateGetItemsQuery(filter)

	log.DebugContext(ctx, "GetItems SQL",
		slog.String("query", query),
		slog.Any("args", args),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "GetItems ошибка бд", slog.Any("err", err), slog.String("store_id", filter.StoreID))
		return nil, err
	}
	defer rows.Close()

	var items []*domain.ItemAgg
	for rows.Next() {
		var item domain.ItemAgg
		var typeIDsJSON string

		err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.Description,
			&item.CardImg,
			&typeIDsJSON,
		)
		if err != nil {
			log.ErrorContext(ctx, "GetItems ошибка при декодировании данных", slog.Any("err", err))
			return nil, err
		}

		var typeIDs []string
		_ = json.Unmarshal([]byte(typeIDsJSON), &typeIDs)

		item.TypesID = typeIDs
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetItems ошибка после чтения строк", slog.Any("err", err), slog.String("store_id", filter.StoreID))
		return nil, err
	}

	if len(items) == 0 {
		log.DebugContext(ctx, "GetItems пустой ответ", slog.String("store_id", filter.StoreID))
		return nil, domain.ErrRowsNotFound
	}

	log.DebugContext(ctx, "GetItems завершено успешно",
		slog.String("store_id", filter.StoreID),
		slog.Int("items_count", len(items)))
	return items, nil
}
