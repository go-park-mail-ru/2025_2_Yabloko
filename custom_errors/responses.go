package custom_errors

import "errors"

var InvalidJSONErr = errors.New("неверные параметры запроса")
var InnerErr = errors.New("неизвестная ошибка сервера")
var HTTPMethodErr = errors.New("метод недоступен")

var EmailAlreadyExist = errors.New("пользователь уже существует")
var IncorrectLoginOrPassword = errors.New("некорректные данные")
var AuthentificationRequired = errors.New("необходима аутентификация")

var TokenExpiredErr = errors.New("токен просрочен")
var TokenTooOldToRefreshErr = errors.New("токен слишком стар для обновления")
var InvalidTokenErr = errors.New("недействительный токен")
var InvalidTokenClaimsErr = errors.New("недействительные данные токена")
