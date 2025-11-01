package domain

import "errors"

var (
	ErrRequestParams  = errors.New("неверные параметры запроса")
	ErrInternalServer = errors.New("неизвестная ошибка сервера")
	ErrHTTPMethod     = errors.New("метод недоступен")
	ErrUnauthorized   = errors.New("ошибка аутентификации")
	ErrForbidden      = errors.New("доступ запрещен")

	ErrRowsNotFound = errors.New("не найдено записей")
	ErrStoreExist   = errors.New("магазин уже существует")
)
