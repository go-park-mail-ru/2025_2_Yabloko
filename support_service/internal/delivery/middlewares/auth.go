package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	guestIDKey  contextKey = "guestID"
	userRoleKey contextKey = "userRole"
)

// AuthMiddleware - обработка JWT и Guest-ID строго по документации
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Обработка JWT токена (Bearer authentication)
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Парсим JWT и извлекаем данные
			userID, role, err := parseJWT(token)
			if err == nil {
				// Успешно распарсили JWT
				if userID != "" {
					ctx = context.WithValue(ctx, userIDKey, userID)
				}
				if role != "" {
					ctx = context.WithValue(ctx, userRoleKey, role)
				}
			}
			// Если JWT невалидный - игнорируем, но не прерываем запрос
			// (возможно, есть guest-id)
		}

		if guestID := r.Header.Get("X-Guest-ID"); guestID != "" {
			ctx = context.WithValue(ctx, guestIDKey, guestID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware - проверка роли администратора строго по документации
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value(userRoleKey).(string)
		if !ok || userRole != "admin" {
			http.Error(w, `{"error": "Не админ"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Вспомогательные функции для извлечения данных из контекста
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func GuestIDFromContext(ctx context.Context) (string, bool) {
	guestID, ok := ctx.Value(guestIDKey).(string)
	return guestID, ok
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok
}

// parseJWT - парсинг JWT токена (заглушка - реализуйте свою логику)
func parseJWT(token string) (userID string, role string, err error) {
	// TODO: Реализовать парсинг JWT с проверкой подписи
	// Пример логики:
	// - Для пользователей: извлекаем user_id из claims
	// - Для админов: проверяем роль "admin" в claims
	// - Возвращаем ошибку если токен невалидный

	// Заглушка для примера:
	if token == "admin_token" {
		return "admin-user-id", "admin", nil
	}
	if token == "user_token" {
		return "user-id", "user", nil
	}

	// Если токен не распознан, возвращаем ошибку
	return "", "", fmt.Errorf("invalid token")
}
