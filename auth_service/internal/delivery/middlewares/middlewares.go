package middlewares

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var CSRFSecret = []byte(getenv("CSRF_SECRET", ""))

func init() {
	if len(CSRFSecret) == 0 {
		slog.Error("CSRF_SECRET is empty or not set")
		panic("CSRF_SECRET required")
	}
	if len(CSRFSecret) < 32 {
		slog.Warn("CSRF_SECRET is too short (min 32 bytes recommended)")
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type ctxKey string

const sessionIDKey ctxKey = "session_id"

type CSRFClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func GenerateJWTCSRFToken(sessionID string) (string, error) {
	claims := CSRFClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(CSRFSecret)
}

func verifyJWTCSRFToken(tokenString, sessionID string) bool {
	token, err := jwt.ParseWithClaims(tokenString, &CSRFClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return CSRFSecret, nil
	})
	if err != nil {
		return false
	}
	if claims, ok := token.Claims.(*CSRFClaims); ok && token.Valid {
		return claims.SessionID == sessionID
	}
	return false
}

func GetSessionID(r *http.Request) string {
	if v := r.Context().Value(sessionIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func AccessLog(log logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		ctx := trace.SetRequestID(r.Context(), reqID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-ID", reqID)

		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()

		log.Info("request started",
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
		)
		next.ServeHTTP(sw, r)
		log.Info("request completed",
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
			slog.Int("status", sw.status),
			slog.Int("bytes", sw.bytes),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sessionID string
		if cookie, err := r.Cookie("session_id"); err == nil && cookie.Value != "" {
			sessionID = cookie.Value
		}
		if sessionID == "" {
			sessionID = uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true"),
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400,
			})
		}
		ctx := context.WithValue(r.Context(), sessionIDKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SmartCSRFMiddleware(next http.Handler) http.Handler {
	skipPaths := map[string]bool{
		"/api/v0/csrf": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		sessionID := GetSessionID(r)
		if sessionID == "" {
			http.Error(w, "Session required", http.StatusForbidden)
			return
		}

		clientToken := r.Header.Get("X-CSRF-Token")
		if clientToken == "" {
			http.Error(w, "CSRF token required in X-CSRF-Token header", http.StatusForbidden)
			return
		}

		if !verifyJWTCSRFToken(clientToken, sessionID) {
			http.Error(w, "Invalid or expired CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
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
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-CSRF-Token, X-Request-ID")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
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

func RateLimit(max int, window time.Duration) func(http.Handler) http.Handler {
	type bucket struct {
		tokens int
		reset  time.Time
	}
	var mu sync.Mutex
	store := make(map[string]*bucket)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}

			now := time.Now()
			mu.Lock()
			b, ok := store[ip]
			if !ok || now.After(b.reset) {
				b = &bucket{tokens: max, reset: now.Add(window)}
				store[ip] = b
			}
			if b.tokens <= 0 {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			b.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
