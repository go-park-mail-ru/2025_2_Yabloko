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
	name        string
	description string
	cityId      uuid.UUID
	address     string
	cardImg     string
	rating      float64
	openAt      time.Time
	closedAt    time.Time
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

	_, err = conn.Exec(ctxt, addStore, store)
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
	Id          string
	name        string
	description string
	city        string
	address     string
	cardImg     string
	rating      float64
	openAt      time.Time
	closedAt    time.Time
}

// todo сортировка и фильтрация
func GetStores(dbPool *pgxpool.Pool, params GetRequest) ([]ResponseInfo, error) {
	logger.Debug(log.LogInfo{Info: "get store", Meta: params})
	ctxt := context.Background()
	var query string

	if params.Sorted == "" && params.TagId == "" {
		if params.LastId == "" {
			query = getStore
		} else {
			query = getStoreFirst
		}
	}

	conn, err := dbPool.Acquire(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка получения коннектора"})
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctxt, query, params.Limit, params.LastId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(log.LogInfo{Err: custom_errors.NotExistErr, Meta: params})
			return nil, custom_errors.NotExistErr
		}
		return nil, err
	}
	defer rows.Close()

	stores := []ResponseInfo{}
	for rows.Next() {
		var store ResponseInfo
		if err := rows.Scan(&store); err != nil {
			logger.Warn(log.LogInfo{Info: "get user password прервано", Err: err, Meta: params})
			return stores, err
		}
		stores = append(stores, store)
	}
	if err = rows.Err(); err != nil {
		logger.Warn(log.LogInfo{Info: "get user password завершено с ошибкой", Err: err, Meta: params})
		return stores, err
	}

	logger.Debug(log.LogInfo{Info: "get user password завершено", Meta: params})
	return stores, nil
}
