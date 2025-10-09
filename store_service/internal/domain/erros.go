package domain

import "errors"

var (
	ErrRequestParams  = errors.New("неверные параметры запроса")
	ErrInternalServer = errors.New("неизвестная ошибка сервера")
	ErrHTTPMethod     = errors.New("метод недоступен")

	ErrStoreNotFound = errors.New("магазин не найден")
	ErrStoreExist    = errors.New("магазин уже существует")
)
