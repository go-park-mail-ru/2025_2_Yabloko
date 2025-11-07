package http

import (
	"apple_backend/auth_service/internal/delivery/middlewares"
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

func NewAuthHandler(uc AuthUseCaseInterface, log logger.Logger) *AuthHandler {
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
	now := time.Now()
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  now.Add(-time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  now.Add(-time.Hour),
		Path:     "/",
		HttpOnly: false,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "Register", domain.ErrHTTPMethod, nil)
		return
	}
	if ct := strings.ToLower(r.Header.Get("Content-Type")); !strings.HasPrefix(ct, "application/json") {
		h.rs.Error(r.Context(), w, http.StatusUnsupportedMediaType, "Register", domain.ErrRequestParams, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.rs.Error(r.Context(), w, http.StatusMethodNotAllowed, "Logout", domain.ErrHTTPMethod, nil)
		return
	}

	clearAuthCookie(w)
	h.rs.Send(r.Context(), w, http.StatusOK, map[string]string{"message": "logged out"})
}

func NewAuthRouter(mux *http.ServeMux, base string, appLog logger.Logger, uc AuthUseCaseInterface) {
	h := NewAuthHandler(uc, appLog)

	mux.Handle(base+"/signup", rateLimitHandler(h.Register))
	mux.Handle(base+"/login", rateLimitHandler(h.Login))
	mux.Handle(base+"/refresh", rateLimitHandler(h.RefreshToken))
	mux.Handle(base+"/logout", rateLimitHandler(h.Logout))
}

func rateLimitHandler(fn http.HandlerFunc) http.Handler {
	return middlewares.RateLimit(5, time.Minute)(fn)
}
