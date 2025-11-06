package middlewares

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

const JwtCookieName = "jwt_token"
const UserIDKey = "user_id"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.Handler, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(JwtCookieName)
		if err != nil {
			http.Error(w, `{"error": "необходима авторизация"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error": "ошибка аутентификации"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000,http://127.0.0.1:3000"
	}
	allowed := parseAllowedOrigins(origins)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqOrigin := r.Header.Get("Origin")
		if reqOrigin != "" && allowed[reqOrigin] {
			w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-CSRF-Token")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func parseAllowedOrigins(v string) map[string]bool {
	m := map[string]bool{}
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			m[s] = true
		}
	}
	return m
}
