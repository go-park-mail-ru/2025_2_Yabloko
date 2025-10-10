package domain

import "errors"

var (
	ErrRequestParams  = errors.New("неверные параметры запроса")
	ErrInternalServer = errors.New("неизвестная ошибка сервера")
	ErrHTTPMethod     = errors.New("метод недоступен")

	ErrRowsNotFound = errors.New("не найдено записей")
	ErrStoreExist   = errors.New("магазин уже существует")
)
