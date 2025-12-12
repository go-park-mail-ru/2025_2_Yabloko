package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/store_service/internal/domain"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"

	"github.com/goccy/go-json"

	"github.com/pgvector/pgvector-go"
)

//go:embed sql/item/get_types.sql
var getItemTypesQuery string

//go:embed sql/item/get_items.sql
var getItemsQuery string

//go:embed sql/item/update_embedding.sql
var updateItemEmbeddingQuery string

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
	log.DebugContext(ctx, "GetItemTypes",
		slog.String("store_id", storeID),
	)

	rows, err := r.db.Query(ctx, getItemTypesQuery, storeID)
	if err != nil {
		log.ErrorContext(ctx, "GetItemTypes query failed",
			slog.Any("err", err),
			slog.String("store_id", storeID),
		)
		return nil, err
	}
	defer rows.Close()

	var itemTypes []*domain.ItemType
	for rows.Next() {
		var itemType domain.ItemType
		err = rows.Scan(&itemType.ID, &itemType.Name)
		if err != nil {
			log.ErrorContext(ctx, "GetItemTypes scan error", slog.Any("err", err))
			return nil, err
		}
		itemTypes = append(itemTypes, &itemType)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetItemTypes rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetItemTypes completed",
		slog.String("store_id", storeID),
		slog.Int("types_count", len(itemTypes)),
	)
	return itemTypes, nil
}

// Для построения фильтров
type ItemQueryBuilder struct {
	whereConditions []string
	args            []any
	paramCount      int
}

func NewItemQueryBuilder() *ItemQueryBuilder {
	return &ItemQueryBuilder{
		whereConditions: []string{},
		args:            []any{},
		paramCount:      0,
	}
}

func (qb *ItemQueryBuilder) AddCondition(condition string, args ...any) {
	for _, arg := range args {
		qb.paramCount++
		qb.args = append(qb.args, arg)
	}
	placeholderCondition := condition
	for i := 1; i <= len(args); i++ {
		placeholderCondition = strings.Replace(
			placeholderCondition,
			fmt.Sprintf("$%d", i),
			fmt.Sprintf("$%d", qb.paramCount-len(args)+i),
			1,
		)
	}
	qb.whereConditions = append(qb.whereConditions, placeholderCondition)
}

func (qb *ItemQueryBuilder) BuildWhere() (string, []any) {
	if len(qb.whereConditions) == 0 {
		return "", qb.args
	}
	return " WHERE " + strings.Join(qb.whereConditions, " AND "), qb.args
}

func (r *ItemRepoPostgres) GetItems(ctx context.Context, filter *domain.ItemFilter) ([]*domain.ItemAgg, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetItems",
		slog.String("store_id", filter.StoreID),
		slog.Any("item_types", filter.ItemTypes),
		slog.String("sorted", filter.Sorted),
		slog.Bool("desc", filter.Desc),
	)

	qb := NewItemQueryBuilder()

	qb.AddCondition(`si.store_id = $1`, filter.StoreID)

	if len(filter.ItemTypes) > 0 {
		placeholders := make([]string, len(filter.ItemTypes))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", qb.paramCount+i+1)
		}
		qb.whereConditions = append(
			qb.whereConditions,
			fmt.Sprintf(`EXISTS (SELECT 1 FROM item_type it2 WHERE it2.item_id = i.id AND it2.type_id = ANY(ARRAY[%s]::uuid[]))`,
				strings.Join(placeholders, ",")),
		)
		for _, typeID := range filter.ItemTypes {
			qb.args = append(qb.args, typeID)
		}
		qb.paramCount += len(filter.ItemTypes)
	}

	where, args := qb.BuildWhere()

	query := getItemsQuery + where + `
	GROUP BY
    si.id,
    i.name,
    si.price,
    i.description,
    i.card_img`

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

	query += "\nORDER BY " + strings.Join(orderClauses, ", ")

	log.DebugContext(ctx, "GetItems final query",
		slog.String("query", query),
		slog.Any("args", args),
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.ErrorContext(ctx, "GetItems query failed",
			slog.Any("err", err),
			slog.String("store_id", filter.StoreID),
		)
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
			log.ErrorContext(ctx, "GetItems scan error", slog.Any("err", err))
			return nil, err
		}

		var typeIDs []string
		if typeIDsJSON == "" || typeIDsJSON == "null" || typeIDsJSON == "[]" {
			typeIDs = []string{}
		} else {
			if err := json.Unmarshal([]byte(typeIDsJSON), &typeIDs); err != nil {
				log.ErrorContext(ctx, "GetItems json unmarshal error",
					slog.Any("err", err),
					slog.String("json", typeIDsJSON))
				return nil, err
			}
		}

		if typeIDs == nil {
			typeIDs = []string{}
		}
		item.TypesID = typeIDs
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "GetItems rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetItems completed",
		slog.String("store_id", filter.StoreID),
		slog.Int("items_count", len(items)),
	)

	if len(items) == 0 {
		return []*domain.ItemAgg{}, nil
	}

	return items, nil
}

func (r *ItemRepoPostgres) UpdateItemEmbedding(ctx context.Context, itemID string, embedding []float32) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "UpdateItemEmbedding", slog.String("id", itemID))

	vec := pgvector.NewVector(embedding)
	_, err := r.db.Exec(ctx, updateItemEmbeddingQuery, itemID, vec)
	if err != nil {
		log.ErrorContext(ctx, "UpdateItemEmbedding query failed", slog.Any("err", err))
		return err
	}

	log.DebugContext(ctx, "UpdateItemEmbedding completed", slog.String("id", itemID))
	return nil
}

//go:embed sql/item/get_items_without_embedding.sql
var getItemsWithoutEmbeddingQuery string

func (r *ItemRepoPostgres) GetItemsWithoutEmbedding(ctx context.Context) ([]*domain.ItemForEmbedding, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "GetItemsWithoutEmbedding")

	rows, err := r.db.Query(ctx, getItemsWithoutEmbeddingQuery)
	if err != nil {
		log.ErrorContext(ctx, "query failed", slog.Any("err", err))
		return nil, err
	}
	defer rows.Close()

	var items []*domain.ItemForEmbedding
	for rows.Next() {
		var item domain.ItemForEmbedding
		err = rows.Scan(&item.ID, &item.Name, &item.Description)
		if err != nil {
			log.ErrorContext(ctx, "scan error", slog.Any("err", err))
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.ErrorContext(ctx, "rows error", slog.Any("err", err))
		return nil, err
	}

	log.DebugContext(ctx, "GetItemsWithoutEmbedding completed", slog.Int("count", len(items)))
	return items, nil
}
