package db

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

var logger log.Logger

func init() {
	const logFile = "./log/bd.log"
	logger = *log.NewLogger("DB-QUERIES", logFile, log.DEBUG)
}

// AppendUser добавляет пользователя
func AppendUser(dbPool PoolDB, email, hashPassword string) (string, error) {

	addUser := `
		insert into account (id, email, hash)
		values ($1, $2, $3);
	`

	logger.Debug(log.LogInfo{Info: "create user", Meta: []string{email}})
	ctxt := context.Background()

	uid := uuid.New().String()
	_, err := dbPool.Exec(ctxt, addUser, uid, email, hashPassword)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			logger.Warn(log.LogInfo{Info: "email уже существует", Err: err, Meta: []string{email}})
			return "", custom_errors.AlreadyExistErr
		}
		logger.Error(log.LogInfo{Err: err, Info: "create user ошибка в процессе запроса"})
		return "", err
	}

	logger.Debug(log.LogInfo{Info: "create user завершено", Meta: []string{email}})
	return uid, nil
}

// DeleteUser удаляет пользователя по email
func DeleteUser(dbPool PoolDB, email string) error {

	deleteUser := `
		delete from account
		where email = $1;
	`

	logger.Debug(log.LogInfo{Info: "delete user", Meta: []string{email}})
	ctxt := context.Background()

	_, err := dbPool.Exec(ctxt, deleteUser, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn(log.LogInfo{Info: "delete user не найден аккаунт",
				Err: custom_errors.NotExistErr, Meta: []string{email}})
		} else {
			logger.Error(log.LogInfo{Err: err, Info: "delete user ошибка в процессе запроса"})
			return err
		}
	}

	logger.Debug(log.LogInfo{Info: "delete user завершено", Meta: []string{email}})
	return nil
}

// GetUserInfo возвращаеь id и hash пароля по email
func GetUserInfo(dbPool PoolDB, email string) (string, string, error) {
	getUser := `
		select id, hash 
		from account
		where email = $1;
	`

	logger.Debug(log.LogInfo{Info: "get user password", Meta: []string{email}})
	ctxt := context.Background()

	var hash, id string
	err := dbPool.QueryRow(ctxt, getUser, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn(log.LogInfo{Err: custom_errors.NotExistErr, Meta: []string{email}})
			return "", "", custom_errors.NotExistErr
		}
		logger.Error(log.LogInfo{Err: err, Meta: []string{email}})
		return "", "", err
	}

	logger.Debug(log.LogInfo{Info: "get user password завершено", Meta: []string{email}})
	return id, hash, nil
}
