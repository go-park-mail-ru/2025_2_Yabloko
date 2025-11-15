package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("пользователь с таким email уже зарегистрирован")
	ErrInvalidEmail      = errors.New("введен некорректный адрес электронной почты")
	ErrWeakPassword      = errors.New("пароль не соответствует требованиям безопасности")
	ErrUserNotFound      = errors.New("учетная запись не найдена")
	ErrInvalidPassword   = errors.New("неверно указан пароль")
	ErrInvalidToken      = errors.New("токен авторизации недействителен")
	ErrHTTPMethod        = errors.New("используемый HTTP-метод не разрешен")
	ErrRequestParams     = errors.New("переданы некорректные параметры запроса")
	ErrInternalServer    = errors.New("произошла внутренняя ошибка сервера")
	ErrUnauthorized      = errors.New("требуется авторизация для доступа к ресурсу")
)
