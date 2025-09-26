package auth

import "github.com/golang-jwt/jwt/v4"

type User struct {
	ID       int    `json:"id"`
	Emain    string `json:"email"`
	Password string `json:"password_hash"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
