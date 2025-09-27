package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// обрабатывает запрос на регистрацию
func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// читаем JSON из тела запроса
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// если JSON некорректный - ошибка 400
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// валидируем поля запроса
	if err := ValidateRegisterRequest(req); err != nil {
		// если валидация не пройдена - ошибка 400
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// проверяем, существует ли пользователь с таким логином
	existingUser, err := findUserByLogin(req.Login)
	if err != nil {
		// ошибка чтения данных
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		return
	}

	if existingUser != nil {
		// если пользователь уже есть - ошибка 409
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
		return
	}

	// хэшируем пароль
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to hash password"})
		return
	}

	// создаем нового пользователя
	if err := createUser(req.Login, passwordHash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
		return
	}

	// возвращаем успешный ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created",
		"login":   req.Login,
	})
}

// обрабатывает запрос на вход
func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// читаем JSON из тела запроса
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// некорректный JSON
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// проверяем данные на валидность
	if err := ValidateLoginRequest(req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// ищем пользователя по логину
	user, err := findUserByLogin(req.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
		return
	}

	// проверяем пароль
	if user == nil || !checkPasswordHash(req.Password, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return
	}

	// создаем JWT токен
	token, err := generateJWT(user.ID, user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate token"})
		return
	}

	// устанавливаем cookie с токеном
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // локально false
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	// возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"login":   req.Login,
	})
}

// refreshTokenHandler обновляет JWT токен
func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// получаем токен из cookie
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// парсим токен
	token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return GetConfig().JWTSecret, nil
	})

	var ve *jwt.ValidationError
	if err != nil && !(errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return
	}

	if token == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return
	}

	// получаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token claims"})
		return
	}

	// проверяем срок жизни токена для refresh
	if claims.ExpiresAt == nil || time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token refresh window expired"})
		return
	}

	if claims.IssuedAt == nil || time.Since(claims.IssuedAt.Time) > 30*24*time.Hour {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token too old to refresh"})
		return
	}

	// создаем новый токен
	newToken, err := generateJWT(claims.UserID, claims.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate token"})
		return
	}

	// устанавливаем новый cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    newToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // локально false
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	// возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Token refreshed",
		"login":   claims.Login,
	})
}

// обрабатывает запрос для выхода
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// обнуляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   false, // локально false
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	// возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out",
	})
}
