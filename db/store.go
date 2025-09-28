package db

import (
	"apple_backend/custom_errors"
	log "apple_backend/logger"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolDB interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Ping(context.Context) error
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	Close()
	Stat() *pgxpool.Stat
	Reset()
	Config() *pgxpool.Config
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type AppendInfo struct {
	Id          uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CityId      uuid.UUID `json:"city_id"`
	Address     string    `json:"address"`
	CardImg     string    `json:"card_img"`
	Rating      float64   `json:"rating"`
	OpenAt      time.Time `json:"open_at"`
	ClosedAt    time.Time `json:"closed_at"`
}

func AppendStore(dbPool PoolDB, store AppendInfo) error {
	addStore := `
		insert into store (id, name, description, city_id, address, card_img, rating, open_at, closed_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`

	logger.Debug(log.LogInfo{Info: "create store", Meta: store})
	ctxt := context.Background()

	if store.Id == uuid.Nil {
		store.Id = uuid.New()
	}
	_, err := dbPool.Exec(ctxt, addStore, store.Id, store.Name, store.Description,
		store.CityId, store.Address, store.CardImg, store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			logger.Warn(log.LogInfo{Err: err, Meta: store})
			return custom_errors.AlreadyExistErr
		}
		logger.Error(log.LogInfo{Err: err, Info: "create store ошибка в процессе запроса", Meta: store})
		return err
	}

	logger.Debug(log.LogInfo{Info: "create store завершено", Meta: store})
	return nil
}

type GetRequest struct {
	Limit  int    `json:"limit"`
	LastId string `json:"last_id"`
	//для сортировки и фильтрации
	TagId  string `json:"tag_id"`
	Sorted string `json:"sorted"`
	Desc   bool   `json:"desc"`
}

type ResponseInfo struct {
	Id          string  `json:"store_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CityID      string  `json:"cit_id"`
	Address     string  `json:"address"`
	CardImg     string  `json:"card_img"`
	Rating      float64 `json:"rating"`
	OpenAt      string  `json:"open_at"`
	ClosedAt    string  `json:"close_at"`
}

func generateQuery(params GetRequest) (string, []any) {
	query := `
		select id, name, description, city_id, address, card_img, rating, open_at, closed_at 
		from store
	`
	where := []string{}
	args := []any{}

	// фильтрация по тегу
	if params.TagId != "" {
		where = append(where, fmt.Sprintf("tag_id = $%d", len(args)+1))
		args = append(args, params.TagId)
	}
	// если не первый запрос
	if params.LastId != "" {
		where = append(where, fmt.Sprintf("id > $%d", len(args)+1))
		args = append(args, params.LastId)
	}

	if len(where) > 0 {
		query += " where " + strings.Join(where, " and ")
	}

	// сортировка
	allowedSort := map[string]string{
		"name":     "name",
		"rating":   "rating",
		"close_at": "close_at",
		"open_at":  "open_at",
	}

	orderBy := " order by id"
	if params.Sorted != "" {
		if col, ok := allowedSort[params.Sorted]; ok {
			dir := "asc"
			if params.Desc {
				dir = "desc"
			}
			orderBy = fmt.Sprintf(" order by %s %s, id", col, dir)
		}
	}
	query += orderBy

	query += fmt.Sprintf(" limit $%d", len(args)+1)
	args = append(args, params.Limit)
	return query, args
}

func GetStores(dbPool PoolDB, params GetRequest) ([]ResponseInfo, error) {
	query, args := generateQuery(params)

	logger.Debug(log.LogInfo{Info: "get store", Meta: params})
	ctxt := context.Background()

	rows, err := dbPool.Query(ctxt, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn(log.LogInfo{Err: custom_errors.NotExistErr, Meta: params})
			return nil, custom_errors.NotExistErr
		}
		logger.Error(log.LogInfo{Err: err, Meta: params})
		return nil, err
	}
	defer rows.Close()

	stores := []ResponseInfo{}
	for rows.Next() {
		var store ResponseInfo
		if err := rows.Scan(&store.Id, &store.Name, &store.Description,
			&store.CityID, &store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt); err != nil {
			logger.Error(log.LogInfo{Info: "get store частично завершено с ошибкой", Err: err, Meta: params})
			return stores, err
		}
		stores = append(stores, store)
	}
	if err = rows.Err(); err != nil {
		logger.Error(log.LogInfo{Info: "get store завершено с ошибкой", Err: err, Meta: params})
		return stores, err
	}

	logger.Debug(log.LogInfo{Info: "get store завершено", Meta: params})
	return stores, nil
}
