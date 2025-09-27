package store

import (
	"apple_backend/custom_errors"
	log "apple_backend/logger"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger log.WebLogger

func init() {
	const logFile = "./log/bd.log"
	logger = log.NewLogger("DB-STORE", logFile, log.DEBUG)
}

type AppendInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CityId      uuid.UUID `json:"cityId"`
	Address     string    `json:"address"`
	CardImg     string    `json:"cardImg"`
	Rating      float64   `json:"rating"`
	OpenAt      time.Time `json:"openAt"`
	ClosedAt    time.Time `json:"closedAt"`
}

func AppendStore(dbPool *pgxpool.Pool, store AppendInfo) error {
	logger.Debug(log.LogInfo{Info: "create store", Meta: store})
	ctxt := context.Background()

	conn, err := dbPool.Acquire(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка получения коннектора"})
		return err
	}
	defer conn.Release()

	begin, err := conn.Begin(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка начала транзакции"})
		return err
	}

	id := uuid.New().String()
	_, err = conn.Exec(ctxt, addStore, id, store.Name, store.Description, store.CityId, store.Address, store.CardImg,
		store.Rating, store.OpenAt, store.ClosedAt)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			logger.Error(log.LogInfo{Err: err, Meta: store})
			return custom_errors.AlreadyExistErr
		}
		logger.Error(log.LogInfo{Err: err, Info: "ошибка в процессе запроса", Meta: store})

		trErr := begin.Rollback(ctxt)
		if trErr != nil {
			logger.Error(log.LogInfo{Err: trErr, Info: "ошибка отката транзакции"})
			return trErr
		}
	}

	err = begin.Commit(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка завершения транзакции"})
		return err
	}

	logger.Debug(log.LogInfo{Info: "create user завершено", Meta: store})
	return nil
}

type GetRequest struct {
	Limit  int    `json:"limit"`
	LastId string `json:"last_id"`
	//для сортировки и фильтрации
	TagId  string `json:"tag_id"`
	Sorted string `json:"sorted"`
}

type ResponseInfo struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	City        string  `json:"city"`
	Address     string  `json:"address"`
	CardImg     string  `json:"card_img"`
	Rating      float64 `json:"rating"`
	OpenAt      string  `json:"open_at"`
	ClosedAt    string  `json:"close_at"`
}

// todo сортировка и фильтрация
func GetStores(dbPool *pgxpool.Pool, params GetRequest) ([]ResponseInfo, error) {
	logger.Debug(log.LogInfo{Info: "get store", Meta: params})
	ctxt := context.Background()
	var query string

	if params.Sorted == "" && params.TagId == "" {
		if params.LastId == "" {
			query = getStoreFirst
		} else {
			query = getStore
		}
	}

	conn, err := dbPool.Acquire(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка получения коннектора"})
		return nil, err
	}
	defer conn.Release()
	//TODO refactor
	var rows pgx.Rows
	if params.LastId == "" {
		rows, err = conn.Query(ctxt, query, params.Limit)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				logger.Error(log.LogInfo{Err: custom_errors.NotExistErr, Meta: params})
				return nil, custom_errors.NotExistErr
			}
			logger.Error(log.LogInfo{Err: err, Meta: params})
			return nil, err
		}
	} else {
		rows, err = conn.Query(ctxt, query, params.Limit, params.LastId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				logger.Error(log.LogInfo{Err: custom_errors.NotExistErr, Meta: params})
				return nil, custom_errors.NotExistErr
			}
			logger.Error(log.LogInfo{Err: err, Meta: params})
			return nil, err
		}
	}

	defer rows.Close()

	stores := []ResponseInfo{}
	for rows.Next() {
		var store ResponseInfo
		if err := rows.Scan(&store.Id, &store.Name, &store.Description, &store.City,
			&store.Address, &store.CardImg, &store.Rating, &store.OpenAt, &store.ClosedAt); err != nil {
			logger.Warn(log.LogInfo{Info: "get store прервано", Err: err, Meta: params})
			return stores, err
		}
		stores = append(stores, store)
	}
	if err = rows.Err(); err != nil {
		logger.Warn(log.LogInfo{Info: "get store завершено с ошибкой", Err: err, Meta: params})
		return stores, err
	}

	logger.Debug(log.LogInfo{Info: "get store завершено", Meta: params})
	return stores, nil
}
