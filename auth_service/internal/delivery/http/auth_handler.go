package http

import (
	authmw "apple_backend/auth_service/internal/delivery/middlewares" // добавлено
	"apple_backend/auth_service/internal/delivery/transport"
	"apple_backend/auth_service/internal/domain"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type AuthUseCaseInterface interface {
	Register(ctx context.Context, email, password string) (*transport.AuthResult, error)
	Login(ctx context.Context, email, password string) (*transport.AuthResult, error)
	RefreshToken(ctx context.Context, tokenString string) (*transport.AuthResult, error)
	VerifyToken(ctx context.Context, tokenString string) (*transport.Claims, error)
	ValidateEmail(ctx context.Context, email string) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}

type AuthHandler struct {
	uc AuthUseCaseInterface
	rs *http_response.ResponseSender
}

func NewAuthHandler(uc AuthUseCaseInterface, log *logger.Logger) *AuthHandler {
	return &AuthHandler{uc: uc, rs: http_response.NewResponseSender(log)}
}

func setAuthCookie(w http.ResponseWriter, token string, expires time.Time) {
	secure := strings.EqualFold(os.Getenv("COOKIE_SECURE"), "true")
	sameSiteStr := strings.ToLower(os.Getenv("COOKIE_SAMESITE"))
	var sameSite http.SameSite = http.SameSiteLaxMode
	if sameSiteStr == "strict" {
		sameSite = http.SameSiteStrictMode
	} else if sameSiteStr == "none" {
		sameSite = http.SameSiteNoneMode
	}
	domain := os.Getenv("COOKIE_DOMAIN")

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Expires:  expires,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Domain:   domain,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body transport.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} transport.AuthResult "Успешная регистрация"
// @Failure 400 {object} http_response.ErrResponse "Некорректные данные"
// @Failure 409 {object} http_response.ErrResponse "Пользователь уже существует"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 415 {object} http_response.ErrResponse "Неверный Content-Type"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /auth/signup [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "Register", domain.ErrHTTPMethod, nil)
		return
	}
	if ct := strings.ToLower(r.Header.Get("Content-Type")); !strings.HasPrefix(ct, "application/json") {
		h.rs.Error(r.Context(), w, http.StatusUnsupportedMediaType, "Register", domain.ErrRequestParams, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
	var req transport.RegisterRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "Register", domain.ErrRequestParams, err)
		return
	}

	res, err := h.uc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			h.rs.Error(r.Context(), w, http.StatusConflict, "Register", err, nil)
		case domain.ErrInvalidEmail, domain.ErrWeakPassword:
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "Register", err, nil)
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "Register", domain.ErrInternalServer, err)
		}
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(r.Context(), w, http.StatusOK, res)
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body transport.LoginRequest true "Данные для входа"
// @Success 200 {object} transport.AuthResult "Успешный вход"
// @Failure 400 {object} http_response.ErrResponse "Некорректные данные"
// @Failure 401 {object} http_response.ErrResponse "Неверные учетные данные"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 415 {object} http_response.ErrResponse "Неверный Content-Type"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "Login", domain.ErrHTTPMethod, nil)
		return
	}
	if ct := strings.ToLower(r.Header.Get("Content-Type")); !strings.HasPrefix(ct, "application/json") {
		h.rs.Error(r.Context(), w, http.StatusUnsupportedMediaType, "Login", domain.ErrRequestParams, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req transport.LoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.rs.Error(r.Context(), w, http.StatusBadRequest, "Login", domain.ErrRequestParams, err)
		return
	}

	res, err := h.uc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound, domain.ErrInvalidPassword:
			h.rs.Error(r.Context(), w, http.StatusUnauthorized, "Login", err, nil)
		case domain.ErrInvalidEmail:
			h.rs.Error(r.Context(), w, http.StatusBadRequest, "Login", err, nil)
		default:
			h.rs.Error(r.Context(), w, http.StatusInternalServerError, "Login", domain.ErrInternalServer, err)
		}
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(r.Context(), w, http.StatusOK, res)
}

// RefreshToken godoc
// @Summary Обновление JWT токена
// @Description Обновляет access token с помощью refresh token из cookies
// @Tags auth
// @Produce json
// @Success 200 {object} transport.AuthResult "Токен успешно обновлен"
// @Failure 401 {object} http_response.ErrResponse "Невалидный или отсутствующий токен"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Failure 500 {object} http_response.ErrResponse "Внутренняя ошибка сервера"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "RefreshToken", domain.ErrHTTPMethod, nil)
		return
	}
	c, err := r.Cookie("jwt_token")
	if err != nil || c.Value == "" {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "RefreshToken", domain.ErrUnauthorized, nil)
		return
	}

	res, err := h.uc.RefreshToken(r.Context(), c.Value)
	if err != nil {
		h.rs.Error(r.Context(), w, http.StatusUnauthorized, "RefreshToken", domain.ErrInvalidToken, err)
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(r.Context(), w, http.StatusOK, res)
}

// Logout godoc
// @Summary Выход из системы
// @Description Очищает JWT токен из cookies
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Успешный выход"
// @Failure 405 {object} http_response.ErrResponse "Неверный HTTP-метод"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "Logout", domain.ErrHTTPMethod, nil)
		return
	}
	clearAuthCookie(w)
	h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{"message": "logged out"})
}

func NewAuthRouter(mux *http.ServeMux, apiPrefix string, appLog *logger.Logger, uc AuthUseCaseInterface) {
	h := NewAuthHandler(uc, appLog)

	base := strings.TrimRight(apiPrefix, "/") + "/auth"

	// 🔧 ДОБАВЬ CSRF middleware ко всем endpoint'ам
	mux.Handle(base+"/signup",
		authmw.CSRFMiddleware(
			authmw.RateLimit(5, time.Minute)(
				http.HandlerFunc(h.Register),
			),
		),
	)
	mux.Handle(base+"/login",
		authmw.CSRFMiddleware(
			authmw.RateLimit(10, time.Minute)(
				http.HandlerFunc(h.Login),
			),
		),
	)
	mux.Handle(base+"/refresh", authmw.CSRFMiddleware(http.HandlerFunc(h.RefreshToken)))
	mux.Handle(base+"/logout", authmw.CSRFMiddleware(http.HandlerFunc(h.Logout)))

	// CSRF endpoint
	mux.Handle(apiPrefix+"/csrf",
		authmw.CSRFTokenMiddleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{
					"message": "CSRF token set in cookies",
				})
			}),
		),
	)
}
