package middlewares

import (
	"apple_backend/auth"
	"context"
	"log"
	"net/http"
	"time"
)

func LoggerWrapper(fun http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		fun(w, r)
		log.Printf("%s %s %s duration %s", r.RemoteAddr, r.Method, r.URL, time.Since(start))
	}
}

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
		claims, err := auth.VerifyJWT(cookie.Value)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// добавляем информацию о пользователе в контекст запроса
		// это позволит последующим обработчикам знать кто делает запрос
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "login", claims.Email)
		r = r.WithContext(ctx)

		// 4. Передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://127.0.0.1:3000": true,
		// production domains...
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-CSRF-Token")
		} else {
			// origin отсутствует (обычный запрос) или не разрешён - не выставляем заголовок
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
