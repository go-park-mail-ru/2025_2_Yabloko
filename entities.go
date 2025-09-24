package main

import "github.com/golang-jwt/jwt/v4"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password_hash"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}
