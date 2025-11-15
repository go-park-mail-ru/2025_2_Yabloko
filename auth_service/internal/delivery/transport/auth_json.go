package transport

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	UserID  string    `json:"user_id"`
	Email   string    `json:"email"`
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}
