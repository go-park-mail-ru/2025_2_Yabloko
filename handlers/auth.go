package handlers

import (
	"apple_backend/auth"
	"apple_backend/custom_errors"
	dbauth "apple_backend/db/auth"
	"apple_backend/logger"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	dbPool *pgxpool.Pool
	log    logger.Logger
}

func NewAuthHandler(dbPool *pgxpool.Pool, logPath string, logLevel logger.LogLevel) *AuthHandler {
	return &AuthHandler{
		dbPool: dbPool,
		log:    *logger.NewLogger("AUTH HANDLER", logPath, logLevel),
	}
}

// структуры для типизированных ответов

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Email   string `json:"email,omitempty"`
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

// TODO переделать под логгер
func handleError(w http.ResponseWriter, statusCode int, userError error, internalErr error) {
	if internalErr != nil {
		log.Printf("Error: %s, internal: %v", userError, internalErr)
	} else {
		log.Printf("Error: %s", userError)
	}

	handleResponse(w, statusCode, ErrorResponse{Error: userError.Error()})
}

func handleSuccess(w http.ResponseWriter, statusCode int, message, email, userID string) {
	response := MessageResponse{
		Message: message,
		Email:   email,
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
		Secure:   false, // fixme: set true in production
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
		Secure:   false, // fixme: set true in production
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}
}

// обрабатывает запрос на регистрацию
func (h *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {

	// читаем JSON из тела запроса
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}

	// валидируем поля запроса
	if err := auth.ValidateRegisterRequest(req); err != nil {
		h.log.Error(logger.LogInfo{Err: err, Info: "неверные поля запроса"})
		handleError(w, http.StatusBadRequest, err, nil)
		return
	}

	// хэшируем пароль
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.log.Error(logger.LogInfo{Err: err, Info: "ошибка хеширования пароля"})
		handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	id, err := dbauth.AppendUser(h.dbPool, req.Email, passwordHash)
	if err != nil {
		if errors.Is(err, custom_errors.AlreadyExistErr) {
			handleError(w, http.StatusConflict, custom_errors.EmailAlreadyExist, nil)
		} else {
			handleError(w, http.StatusConflict, custom_errors.InnerErr, err)
		}
		return
	}
	//TODO
	token, err := auth.GenerateJWT(id, req.Email)
	if err != nil {
		h.log.Warn(logger.LogInfo{Err: err, Info: "аккаунт создан без jwt"})
		handleSuccess(w, http.StatusCreated, "OK без cookie", req.Email, id)
		return
	}

	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusOK, "OK", req.Email, id)
}

// обрабатывает запрос на вход
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, http.StatusBadRequest, custom_errors.InvalidJSONErr, err)
		return
	}

	if err := auth.ValidateLoginRequest(req); err != nil {
		handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, nil)
		return
	}

	id, hashPassword, err := dbauth.GetUserInfo(h.dbPool, req.Email)
	if err != nil {
		if errors.Is(err, custom_errors.NotExistErr) {
			handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, err)
		} else {
			handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		}
		return
	}

	if !auth.CheckPasswordHash(req.Password, hashPassword) {
		handleError(w, http.StatusUnauthorized, custom_errors.IncorrectLoginOrPassword, nil)
		return
	}

	token, err := auth.GenerateJWT(id, req.Email)
	if err != nil {
		handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		h.log.Error(logger.LogInfo{Err: err, Info: "Не удалось создать jwt"})
		return
	}

	http.SetCookie(w, createTokenCookie(token, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusOK, "OK", req.Email, "")
}

// refreshTokenHandler обновляет JWT токен
func (h *AuthHandler) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// получаем токен из cookie
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		handleError(w, http.StatusUnauthorized, custom_errors.AuthentificationRequired, err)
		return
	}

	// парсим токен
	token, err := jwt.ParseWithClaims(cookie.Value, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return auth.JwtSecret, nil
	})

	//todo стандартизировать ошибки и вынести в кустом еррорс
	var ve *jwt.ValidationError
	if err != nil && !(errors.As(err, &ve) && ve.Errors&jwt.ValidationErrorExpired != 0) {
		handleError(w, http.StatusUnauthorized, errors.New("Invalid token"), err)
		return
	}

	if token == nil {
		handleError(w, http.StatusUnauthorized, errors.New("Invalid token"), nil)
		return
	}

	// получаем claims
	claims, ok := token.Claims.(*auth.Claims)
	if !ok {
		handleError(w, http.StatusUnauthorized, errors.New("Invalid token claims"), nil)
		return
	}

	// проверяем срок жизни токена для refresh
	if claims.ExpiresAt == nil || time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
		handleError(w, http.StatusUnauthorized, errors.New("Token expired"), nil)
		return
	}

	if claims.IssuedAt == nil || time.Since(claims.IssuedAt.Time) > 30*24*time.Hour {
		handleError(w, http.StatusUnauthorized, errors.New("Token too old to refresh"), nil)
		return
	}

	newToken, err := auth.GenerateJWT(claims.UserID, claims.Email)
	if err != nil {
		h.log.Error(logger.LogInfo{Err: err, Info: "Ошибка создания jwt"})
		handleError(w, http.StatusInternalServerError, custom_errors.InnerErr, err)
		return
	}

	http.SetCookie(w, createTokenCookie(newToken, time.Now().Add(24*time.Hour)))
	handleSuccess(w, http.StatusOK, "Токен обновлён", claims.Email, "")
}

// обрабатывает запрос для выхода
func (*AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// обнуляем cookie
	http.SetCookie(w, createExpiredTokenCookie())
	handleSuccess(w, http.StatusOK, "Выход выполнен", "", "")
}
