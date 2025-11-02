package domain

import "errors"

var (
	ErrRequestParams  = errors.New("неверные параметры запроса")
	ErrInternalServer = errors.New("неизвестная ошибка сервера")
	ErrHTTPMethod     = errors.New("метод недоступен")

	ErrProfileNotFound    = errors.New("профиль не найден")
	ErrProfileExist       = errors.New("профиль уже существует")
	ErrInvalidProfileData = errors.New("неверные данные профиля")

	ErrFileTooLarge    = errors.New("слишком большой размер файла")
	ErrInvalidFileType = errors.New("недопустимый формат файла")
	ErrUnauthorized    = errors.New("неавторизованный доступ")
	ErrForbidden       = errors.New("доступ запрещен")
)
