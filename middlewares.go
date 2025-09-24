package main

import (
	"context"
	"net/http"
)

// authMiddleware проверяет JWT токен для защищенных routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// извлекаем JWT токен из cookie
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
			return
		}

		// проверяем валидность токена
		claims, err := verifyJWT(cookie.Value)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// добавляем информацию о пользователе в контекст запроса
		// это позволит последующим обработчикам знать кто делает запрос
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "login", claims.Login)
		r = r.WithContext(ctx)

		// 4. Передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	}
}
