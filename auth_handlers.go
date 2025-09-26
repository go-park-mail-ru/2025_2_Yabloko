package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// структуры для типизированных ответов

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Login   string `json:"login,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// хелпер для отправки ответов
func handleResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encoding error: %v", err)
	}
}

func handleError(w http.ResponseWriter, statusCode int, errorMsg string, internalErr error) {
	if internalErr != nil {
		log.Printf("Error: %s, internal: %v", errorMsg, internalErr)
	} else {
		log.Printf("Error: %s", errorMsg)
	}

	handleResponse(w, statusCode, ErrorResponse{Error: errorMsg})
}

func handleSuccess(w http.ResponseWriter, statusCode int, message, login, userID string) {
	response := MessageResponse{
		Message: message,
		Login:   login,
	}
	if userID != "" {
		response.UserID = userID
	}
	handleResponse(w, statusCode, response)
}

func createTokenCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   false, // TODO: set true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
}

func createExpiredTokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   false, // TODO: set true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
}

// обрабатывает запрос на регистрацию
func registerHandler(w http.ResponseWriter, r *http.Request) {

	// читаем JSON из тела запроса
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// валидируем поля запроса
	if err := ValidateRegisterRequest(req); err != nil {
		handleError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	existingUser, err := findUserByLogin(req.Login)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	if existingUser != nil {
		handleError(w, http.StatusConflict, "Пользователь уже существует", nil)
		return
	}

	// хэшируем пароль
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	if err := createUser(req.Login, passwordHash); err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	user, err := findUserByLogin(req.Login)
	if err != nil || user == nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	token, err := generateJWT(user.ID, user.Login)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusCreated, "Пользователь создан и автоматически авторизован", req.Login, fmt.Sprintf("%d", user.ID))
}

// обрабатывает запрос на вход
func loginHandler(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := ValidateLoginRequest(req); err != nil {
		handleError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := findUserByLogin(req.Login)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	if user == nil || !checkPasswordHash(req.Password, user.Password) {
		handleError(w, http.StatusUnauthorized, "Неверный логин или пароль", nil)
		return
	}

	token, err := generateJWT(user.ID, user.Login)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusOK, "Успешный вход", req.Login, "")
}

// refreshTokenHandler обновляет JWT токен
func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// получаем токен из cookie
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		handleError(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	// парсим токен
	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	var ve *jwt.ValidationError
	if err != nil && !(errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0) {
		handleError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	if token == nil {
		handleError(w, http.StatusUnauthorized, "Invalid token", nil)
		return
	}

	// получаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		handleError(w, http.StatusUnauthorized, "Invalid token claims", nil)
		return
	}

	// проверяем срок жизни токена для refresh
	if claims.ExpiresAt == nil || time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
		handleError(w, http.StatusUnauthorized, "Token expired", nil)
		return
	}

	if claims.IssuedAt == nil || time.Since(claims.IssuedAt.Time) > 30*24*time.Hour {
		handleError(w, http.StatusUnauthorized, "Token too old to refresh", nil)
		return
	}

	newToken, err := generateJWT(claims.UserID, claims.Login)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	http.SetCookie(w, createTokenCookie(newToken, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusOK, "Токен обновлён", claims.Login, "")
}

// обрабатывает запрос для выхода
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// обнуляем cookie
	http.SetCookie(w, createExpiredTokenCookie())
	handleSuccess(w, http.StatusOK, "Выход выполнен", "", "")
}
