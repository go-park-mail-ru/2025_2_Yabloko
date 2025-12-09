package middlewares

import (
	"apple_backend/pkg/blacklist"
	"context"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

const JwtCookieName = "jwt_token"

type ctxKey string

var UserIDKey ctxKey = "user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(UserIDKey)
	id, ok := v.(string)
	return id, ok
}

func AuthMiddleware(
	tb blacklist.TokenBlacklist,
	jwtSecret string,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	secret := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(JwtCookieName)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			token := c.Value

			isBlacklisted, err := tb.IsBlacklisted(r.Context(), token)
			if err != nil {
				logger.ErrorContext(r.Context(), "blacklist check failed", slog.Any("err", err))
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
			if isBlacklisted {
				logger.WarnContext(r.Context(), "attempt to use blacklisted token")
				http.Error(w, `{"error":"token revoked"}`, http.StatusUnauthorized)
				return
			}

			type claims struct {
				UserID string `json:"user_id"`
				jwt.RegisteredClaims
			}
			cl := &claims{}
			tkn, err := jwt.ParseWithClaims(token, cl, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})
			if err != nil || !tkn.Valid || cl.UserID == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := WithUserID(r.Context(), cl.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
