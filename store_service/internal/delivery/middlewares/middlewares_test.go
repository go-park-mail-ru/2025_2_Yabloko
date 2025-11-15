package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "secret"

	validToken := func() string {
		claims := &Claims{
			UserID: "user123",
			Email:  "test@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		str, _ := token.SignedString([]byte(secret))
		return str
	}

	expiredToken := func() string {
		claims := &Claims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		str, _ := token.SignedString([]byte(secret))
		return str
	}

	tests := []struct {
		name           string
		cookieValue    string
		setCookie      bool
		expectedStatus int
		expectNextCall bool
	}{
		{
			name:           "валидный токен",
			cookieValue:    validToken(),
			setCookie:      true,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "нет cookie",
			setCookie:      false,
			expectedStatus: http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "невалидный токен",
			cookieValue:    "broken.jwt.token",
			setCookie:      true,
			expectedStatus: http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:           "истёкший токен",
			cookieValue:    expiredToken(),
			setCookie:      true,
			expectedStatus: http.StatusUnauthorized,
			expectNextCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				userID := r.Context().Value(UserIDKey)
				require.Equal(t, "user123", userID)
				w.WriteHeader(http.StatusOK)
			})

			handler := AuthMiddleware(next, secret)

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  JwtCookieName,
					Value: tt.cookieValue,
				})
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			res := rec.Result()

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			require.Equal(t, tt.expectNextCall, nextCalled)
		})
	}
}
