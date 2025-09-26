package auth

import (
	"apple_backend/custom_errors"
	log "apple_backend/logger"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var logger log.WebLogger

func init() {
	const logFile = "./log/bd.log"
	logger = log.NewLogger("DB-AUTH", logFile, log.DEBUG)
}

//TODO подумать как рефакторить обретку транзакций

// AppendUser добавляет пользователя
func AppendUser(dbPool *pgxpool.Pool, email, hashPassword string) (string, error) {
	logger.Debug(log.LogInfo{Info: "create user", Meta: []string{email}})
	ctxt := context.Background()

	conn, err := dbPool.Acquire(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка получения коннектора"})
		return "", err
	}
	defer conn.Release()

	begin, err := conn.Begin(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка начала транзакции"})
		return "", err
	}

	uid := uuid.New().String()
	_, err = conn.Exec(ctxt, addUser, uid, email, hashPassword)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			logger.Error(log.LogInfo{Err: err, Meta: []string{email}})
			return "", custom_errors.AlreadyExistErr
		}
		logger.Error(log.LogInfo{Err: err, Info: "ошибка в процессе запроса"})

		trErr := begin.Rollback(ctxt)
		if trErr != nil {
			logger.Error(log.LogInfo{Err: trErr, Info: "ошибка отката транзакции"})
			return "", trErr
		}
	}

	err = begin.Commit(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка завершения транзакции"})
		return "", err
	}

	logger.Debug(log.LogInfo{Info: "create user завершено", Meta: []string{email}})
	return uid, nil
}

// DeleteUser удаляет пользователя по email
func DeleteUser(dbPool *pgxpool.Pool, email string) error {
	logger.Debug(log.LogInfo{Info: "delete user", Meta: []string{email}})
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

	_, err = conn.Exec(ctxt, deleteUser, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn(log.LogInfo{Info: "delete user не найден аккаунт",
				Err: custom_errors.NotExistErr, Meta: []string{email}})
		} else {
			trErr := begin.Rollback(ctxt)
			if trErr != nil {
				logger.Error(log.LogInfo{Err: trErr, Info: "ошибка отката транзакции"})
				return trErr
			}
			return err
		}
	}

	err = begin.Commit(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка завершения транзакции"})
		return err
	}

	logger.Debug(log.LogInfo{Info: "delete user завершено", Meta: []string{email}})
	return nil
}

type userInfo struct {
	id   string
	hash string
}

// GetUserInfo возвращаеь id и hash пароля по email
func GetUserInfo(dbPool *pgxpool.Pool, email string) (string, string, error) {
	logger.Debug(log.LogInfo{Info: "get user password", Meta: []string{email}})
	ctxt := context.Background()

	conn, err := dbPool.Acquire(ctxt)
	if err != nil {
		logger.Error(log.LogInfo{Err: err, Info: "ошибка получения коннектора"})
		return "", "", err
	}
	defer conn.Release()

	var idHash userInfo
	err = conn.QueryRow(ctxt, getUser, email).Scan(&idHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error(log.LogInfo{Err: custom_errors.NotExistErr, Meta: []string{email}})
			return "", "", custom_errors.NotExistErr
		}
		return "", "", err
	}

	logger.Debug(log.LogInfo{Info: "get user password завершено", Meta: []string{email}})
	return idHash.id, idHash.hash, nil
}
