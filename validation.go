package main

import (
	"regexp"
	"strings"
	"unicode"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// валидация для авторизации
func ValidateLoginRequest(req LoginRequest) error {
	// проверка на пустые поля
	if strings.TrimSpace(req.Login) == "" {
		return &ValidationError{Message: "Login is required"}
	}
	if strings.TrimSpace(req.Password) == "" {
		return &ValidationError{Message: "Password is required"}
	}

	// проверка длины логина
	if len(req.Login) < 3 {
		return &ValidationError{Message: "Login must be at least 3 characters long"}
	}
	if len(req.Login) > 50 {
		return &ValidationError{Message: "Login must be less than 50 characters"}
	}

	// проверка формата логина (только буквы, цифры, подчеркивание)
	if matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", req.Login); !matched {
		return &ValidationError{Message: "Login can only contain letters, numbers and underscores"}
	}

	return nil
}

// валидация для регистрации
func ValidateRegisterRequest(req RegisterRequest) error {
	// валидация логина и базовых проверок
	if err := ValidateLoginRequest(LoginRequest{
		Login:    req.Login,
		Password: req.Password,
	}); err != nil {
		return err
	}

	// дополнительная валидация сложности пароля (только для регистрации)
	if err := ValidatePassword(req.Login, req.Password); err != nil {
		return err
	}

	return nil
}

// проверка сложности пароля
func ValidatePassword(login, password string) error {
	if len(password) < 8 {
		return &ValidationError{Message: "Password must be at least 8 characters long"}
	}

	// проверка на наличие разных символов
	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &ValidationError{Message: "Password must contain at least one uppercase letter"}
	}
	if !hasLower {
		return &ValidationError{Message: "Password must contain at least one lowercase letter"}
	}
	if !hasNumber {
		return &ValidationError{Message: "Password must contain at least one number"}
	}
	if !hasSpecial {
		return &ValidationError{Message: "Password must contain at least one special character"}
	}

	// проверка на несовпадение логина и пароля в приведении к одному регистру
	if strings.EqualFold(login, password) {
		return &ValidationError{Message: "Password cannot be the same as login"}
	}

	return nil
}
