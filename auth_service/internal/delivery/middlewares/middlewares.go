package middlewares

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

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

func AccessLog(baseLogger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// set request id in trace and response header
		ctx := trace.SetRequestID(r.Context(), reqID)
		ctx = logger.ContextWithRequestID(ctx, reqID)
		w.Header().Set("X-Request-Id", reqID)

		// create per-request logger and put into context
		reqLogger := baseLogger.With(
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		// Добавляем логгер в контекст
		ctx = logger.ContextWithLogger(ctx, reqLogger)
		r = r.WithContext(ctx)

		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()

		// Логируем через per-request логгер ИСПОЛЬЗУЯ КОНТЕКСТ
		reqLogger.InfoContext(ctx, "request started")
		next.ServeHTTP(sw, r)

		duration := time.Since(start)
		reqLogger.InfoContext(ctx, "request completed",
			slog.Int("status", sw.status),
			slog.Int("bytes", sw.bytes),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.Float64("duration_seconds", duration.Seconds()))
	})
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		headerToken := r.Header.Get("X-CSRF-Token")
		if headerToken == "" {
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		cookieToken, err := r.Cookie("csrf_token")
		if err != nil || cookieToken.Value == "" {
			http.Error(w, "CSRF cookie missing", http.StatusForbidden)
			return
		}

		if headerToken != cookieToken.Value {
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CSRFTokenMiddleware(next http.Handler) http.Handler {
	secure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true")
	sameSite := http.SameSiteLaxMode
	switch strings.ToLower(os.Getenv("COOKIE_SAMESITE")) {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
		if !secure {
			secure = true
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// csrf_token — для Double Submit (HttpOnly=false)
		if _, err := r.Cookie("csrf_token"); err != nil {
			csrfToken := uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    csrfToken,
				Path:     "/",
				HttpOnly: false,
				Secure:   secure,
				SameSite: sameSite,
				MaxAge:   86400,
			})
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		}

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
			if b.tokens <= 0 && false {
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
