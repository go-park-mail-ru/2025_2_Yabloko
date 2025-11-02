package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ctxKey string

const (
	CtxUserID     ctxKey = "user_id"
	CtxIsAdmin    ctxKey = "is_admin"
	JwtCookieName        = "jwt_token"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if secret == "" {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "server misconfigured: SECRET_KEY is empty", http.StatusInternalServerError)
			})
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tok string

			if c, err := r.Cookie(JwtCookieName); err == nil && c.Value != "" {
				tok = c.Value
			}

			if tok == "" {
				if h := r.Header.Get("Authorization"); h != "" {
					ps := strings.SplitN(h, " ", 2)
					if len(ps) == 2 && strings.EqualFold(ps[0], "Bearer") {
						tok = ps[1]
					}
				}
			}

			if tok == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			t, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !t.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, ok := t.Claims.(jwt.MapClaims)

			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			sub, _ := claims["user_id"].(string)

			if sub == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), CtxUserID, sub)

			if v, ok := claims["is_admin"].(bool); ok && v {
				ctx = context.WithValue(ctx, CtxIsAdmin, true)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JWT() func(http.Handler) http.Handler {
	return AuthMiddleware(os.Getenv("SECRET_KEY"))
}
