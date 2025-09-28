package handlers

import (
	"apple_backend/auth"
	"apple_backend/custom_errors"
	db "apple_backend/db"
	"apple_backend/logger"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// структуры для типизированных ответов

type AuthResponse struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

func createTokenCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   false, // fixme: set true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}

func createExpiredTokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   false, // fixme: set true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}

// обрабатывает запрос на регистрацию
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {

	// читаем JSON из тела запроса
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}

	// валидируем поля запроса
	if err := auth.ValidateRegisterRequest(req); err != nil {
		h.handleError(w, http.StatusBadRequest, custom_errors.IncorrectLoginOrPassword, err)
		return
	}

	// хэшируем пароль
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	id, err := db.AppendUser(h.dbPool, req.Email, passwordHash)
	if err != nil {
		if errors.Is(err, custom_errors.AlreadyExistErr) {
			h.handleError(w, http.StatusConflict, custom_errors.EmailAlreadyExist, nil)
		} else {
			h.handleError(w, http.StatusConflict, custom_errors.InnerErr, err)
		}
		return
	}

	response := AuthResponse{
		UserID: id,
		Email:  req.Email,
	}

	token, err := auth.GenerateJWT(id, req.Email)
	if err != nil {
		h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	response.Message = "OK"
	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	h.handleResponse(w, http.StatusOK, response)
}

// обрабатывает запрос на вход
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}

	if err := auth.ValidateLoginRequest(req); err != nil {
		h.handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, nil)
		return
	}

	id, hashPassword, err := db.GetUserInfo(h.dbPool, req.Email)
	if err != nil {
		if errors.Is(err, custom_errors.NotExistErr) {
			h.handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, err)
		} else {
			h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		}
		return
	}

	if !auth.CheckPasswordHash(req.Password, hashPassword) {
		h.handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, nil)
		return
	}

	token, err := auth.GenerateJWT(id, req.Email)
	if err != nil {
		h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		h.log.Error(logger.LogInfo{Err: err, Info: "Не удалось создать jwt"})
		return
	}

	response := AuthResponse{
		UserID:  id,
		Email:   req.Email,
		Message: "OK",
	}
	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	h.handleResponse(w, http.StatusOK, response)
}

// refreshTokenHandler обновляет JWT токен
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// получаем токен из cookie
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		h.handleError(w, http.StatusUnauthorized, custom_errors.AuthentificationRequired, err)
		return
	}

	// парсим токен
	token, err := jwt.ParseWithClaims(cookie.Value, &auth.Claims{}, auth.ParseJWT)

	//todo стандартизировать ошибки и вынести в кустом еррорс
	var ve *jwt.ValidationError
	if err != nil && !(errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0) {
		h.handleError(w, http.StatusUnauthorized, custom_errors.InvalidTokenErr, err)
		return
	}

	if token == nil {
		h.handleError(w, http.StatusUnauthorized, custom_errors.InvalidTokenErr, nil)
		return
	}

	// получаем claims
	claims, ok := token.Claims.(*auth.Claims)
	if !ok {
		h.handleError(w, http.StatusUnauthorized, custom_errors.InvalidTokenClaimsErr, nil)
		return
	}

	// проверяем срок жизни токена для refresh
	if claims.ExpiresAt == nil || time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
		h.handleError(w, http.StatusUnauthorized, custom_errors.TokenExpiredErr, nil)
		return
	}

	if claims.IssuedAt == nil || time.Since(claims.IssuedAt.Time) > 30*24*time.Hour {
		h.handleError(w, http.StatusUnauthorized, custom_errors.TokenTooOldToRefreshErr, nil)
		return
	}

	newToken, err := auth.GenerateJWT(claims.UserID, claims.Email)
	if err != nil {
		h.log.Error(logger.LogInfo{Err: err, Info: "Ошибка создания jwt"})
		h.handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	response := AuthResponse{
		UserID:  "",
		Email:   claims.Email,
		Message: "OK",
	}
	http.SetCookie(w, createTokenCookie(newToken, time.Now().Add(24*time.Hour)))
	h.handleResponse(w, http.StatusOK, response)
}

// обрабатывает запрос для выхода
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// обнуляем cookie
	http.SetCookie(w, createExpiredTokenCookie())
	h.handleResponse(w, http.StatusOK, nil)
}
