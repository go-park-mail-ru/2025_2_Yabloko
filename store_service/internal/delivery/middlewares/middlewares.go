package middlewares

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func AccessLog(log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := trace.SetRequestID(r.Context(), reqID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", reqID)

		start := time.Now()
		log.Info(ctx, "request started", map[string]interface{}{
			"method": r.Method,
			"url":    r.URL.Path,
		})

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Info(ctx, "request completed", map[string]interface{}{
			"method":   r.Method,
			"url":      r.URL.Path,
			"duration": duration.Milliseconds(),
		})
	})
}

// fixme clean arch request to auth serv
// authMiddleware проверяет JWT токен для защищенных routes
//func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
//
//	return func(w http.ResponseWriter, r *http.Request) {
//		// извлекаем JWT токен из cookie
//		cookie, err := r.Cookie("jwt_token")
//		if err != nil {
//			http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
//			return
//		}
//
//		// проверяем валидность токена
//		claims, err := auth.VerifyJWT(cookie.Value)
//		if err != nil {
//			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
//			return
//		}
//
//		// добавляем информацию о пользователе в контекст запроса
//		// это позволит последующим обработчикам знать кто делает запрос
//		ctx := r.Context()
//		ctx = context.WithValue(ctx, "userID", claims.UserID)
//		ctx = context.WithValue(ctx, "login", claims.Email)
//		r = r.WithContext(ctx)
//
//		// 4. Передаем запрос следующему обработчику
//		next.ServeHTTP(w, r)
//	}
//}

func CorsMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://90.156.218.233")
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-CSRF-Token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
