package middlewares

import (
	"context"
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

func AuthMiddleware(next http.Handler, jwtSecret string) http.Handler {
	secret := []byte(jwtSecret)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(JwtCookieName)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		type claims struct {
			UserID string `json:"user_id"`
			jwt.RegisteredClaims
		}
		cl := &claims{}
		tkn, err := jwt.ParseWithClaims(c.Value, cl, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
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
