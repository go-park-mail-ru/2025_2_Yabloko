package middlewares

import (
	"apple_backend/pkg/logger"
	"apple_backend/pkg/trace"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var CSRFSecret = []byte(getenv("CSRF_SECRET", "dev-csrf-secret"))

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type CSRFClaims struct {
	SessionID string `json:"session_id"`
	UserAgent string `json:"user_agent"`
	jwt.RegisteredClaims
}

func generateJWTCSRFToken(sessionID string, userAgent string) (string, error) {
	claims := CSRFClaims{
		SessionID: sessionID,
		UserAgent: userAgent,
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

func verifyJWTCSRFToken(tokenString, sessionID, userAgent string) bool {
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
		return claims.SessionID == sessionID && claims.UserAgent == userAgent
	}
	return false
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

func AccessLog(log *logger.Logger, next http.Handler) http.Handler {
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

		log.Info(ctx, "request started", map[string]interface{}{"method": r.Method, "url": r.URL.Path})
		next.ServeHTTP(sw, r)
		log.Info(ctx, "request completed", map[string]interface{}{
			"method": r.Method, "url": r.URL.Path, "status": sw.status, "bytes": sw.bytes, "duration_ms": time.Since(start).Milliseconds(),
		})
	})
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		clientToken := r.Header.Get("X-CSRF-Token")
		if clientToken == "" {
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Session required", http.StatusForbidden)
			return
		}
		if !verifyJWTCSRFToken(clientToken, sessionCookie.Value, r.UserAgent()) {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CSRFTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_id")
		var sessionID string
		if err != nil {
			sessionID = uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400,
			})
		} else {
			sessionID = sessionCookie.Value
		}
		if _, err := r.Cookie("csrf_token"); err != nil {
			csrfToken, err := generateJWTCSRFToken(sessionID, r.UserAgent())
			if err != nil {
				http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    csrfToken,
				Path:     "/",
				HttpOnly: false,
				Secure:   false,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400,
			})
		}
		next.ServeHTTP(w, r)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("SERVER_BASE_URL")
		if origin == "" {
			origin = "http://localhost:3000"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
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

// Простой rate limit по IP: N запросов за window
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
