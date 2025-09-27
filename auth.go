package main

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// jwtSecret - секретный ключ для подписи JWT токенов
// TODO: потом перенести это в environment variables (сделать невидимым для всех)

// generateJWT создает новый JWT токен для пользователя
func generateJWT(userID int, login string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()), // ← ДОБАВИТЬ
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(GetConfig().JWTSecret)
}

func verifyJWT(tokenString string) (*Claims, error) {
	// парсим и передаем анонимную функцию как инструкцию
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// проверяем HMAC или нет
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return GetConfig().JWTSecret, nil

	})

	if err != nil {
		return nil, err
	}
	// проверка валидности токена и соответствие claims структуре ожидаемой
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid

}

// создает bcrypt хэш пароля
func hashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10 (оптимальное значение для баланса безопасности и производительности)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// проверяет соответствие пароля и хэша
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
