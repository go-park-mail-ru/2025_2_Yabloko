package custom_errors

import "errors"

var AlreadyExistErr = errors.New("запись с уникальным полем уже существует")
var NotExistErr = errors.New("записи не существует")
