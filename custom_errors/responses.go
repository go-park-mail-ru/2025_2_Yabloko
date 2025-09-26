package custom_errors

import "errors"

var InvalidJSONErr = errors.New("неверные параметры запроса")
var InnerErr = errors.New("неизвестная ошибка сервера")

var EmailAlreadyExist = errors.New("пользователь уже существует")
var IncorrectLoginOrPassword = errors.New("некорректные данные")
var AuthentificationRequired = errors.New("необходима аутентификация")
