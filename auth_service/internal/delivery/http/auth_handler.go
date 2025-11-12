package http

import (
	"apple_backend/auth_service/internal/delivery/middlewares"
	"apple_backend/auth_service/internal/delivery/transport"
	"apple_backend/auth_service/internal/domain"
	"apple_backend/pkg/http_response"
	"apple_backend/pkg/logger"
	"context"
	"encoding/json"
	"log/slog"
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

func NewAuthRouter(mux *http.ServeMux, base string, uc AuthUseCaseInterface) {
	h := NewAuthHandler(uc)

	mux.Handle(base+"/signup", rateLimitHandler(h.Register))
	mux.Handle(base+"/login", rateLimitHandler(h.Login))
	mux.Handle(base+"/refresh", rateLimitHandler(h.RefreshToken))
	mux.Handle(base+"/logout", rateLimitHandler(h.Logout))
}

func NewAuthHandler(uc AuthUseCaseInterface) *AuthHandler {
	return &AuthHandler{
		uc: uc,
		rs: http_response.NewResponseSender(logger.Global()),
	}
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler Register start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler Register wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "Register", domain.ErrHTTPMethod, nil)
		return
	}
	if ct := strings.ToLower(r.Header.Get("Content-Type")); !strings.HasPrefix(ct, "application/json") {
		log.WarnContext(ctx, "handler Register unsupported content type", slog.String("content_type", ct))
		h.rs.Error(ctx, w, http.StatusUnsupportedMediaType, "Register", domain.ErrRequestParams, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req transport.RegisterRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler Register decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "Register", domain.ErrRequestParams, err)
		return
	}

	res, err := h.uc.Register(ctx, req.Email, req.Password)
	if err != nil {
		log.ErrorContext(ctx, "usecase Register failed", slog.Any("err", err))
		switch err {
		case domain.ErrUserAlreadyExists:
			h.rs.Error(ctx, w, http.StatusConflict, "Register", err, nil)
		case domain.ErrInvalidEmail, domain.ErrWeakPassword:
			h.rs.Error(ctx, w, http.StatusBadRequest, "Register", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "Register", domain.ErrInternalServer, err)
		}
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(ctx, w, http.StatusOK, res)
	log.InfoContext(ctx, "handler Register success", slog.String("user_id", res.UserID))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler Login start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler Login wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "Login", domain.ErrHTTPMethod, nil)
		return
	}
	if ct := strings.ToLower(r.Header.Get("Content-Type")); !strings.HasPrefix(ct, "application/json") {
		log.WarnContext(ctx, "handler Login unsupported content type", slog.String("content_type", ct))
		h.rs.Error(ctx, w, http.StatusUnsupportedMediaType, "Login", domain.ErrRequestParams, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req transport.LoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		log.ErrorContext(ctx, "handler Login decode failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusBadRequest, "Login", domain.ErrRequestParams, err)
		return
	}

	res, err := h.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.ErrorContext(ctx, "usecase Login failed", slog.Any("err", err))
		switch err {
		case domain.ErrUserNotFound, domain.ErrInvalidPassword:
			h.rs.Error(ctx, w, http.StatusUnauthorized, "Login", err, nil)
		case domain.ErrInvalidEmail:
			h.rs.Error(ctx, w, http.StatusBadRequest, "Login", err, nil)
		default:
			h.rs.Error(ctx, w, http.StatusInternalServerError, "Login", domain.ErrInternalServer, err)
		}
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(ctx, w, http.StatusOK, res)
	log.InfoContext(ctx, "handler Login success", slog.String("user_id", res.UserID))
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler RefreshToken start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler RefreshToken wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "RefreshToken", domain.ErrHTTPMethod, nil)
		return
	}

	c, err := r.Cookie("jwt_token")
	if err != nil || c.Value == "" {
		log.WarnContext(ctx, "handler RefreshToken missing jwt cookie")
		h.rs.Error(ctx, w, http.StatusUnauthorized, "RefreshToken", domain.ErrUnauthorized, nil)
		return
	}

	res, err := h.uc.RefreshToken(ctx, c.Value)
	if err != nil {
		log.ErrorContext(ctx, "usecase RefreshToken failed", slog.Any("err", err))
		h.rs.Error(ctx, w, http.StatusUnauthorized, "RefreshToken", domain.ErrInvalidToken, err)
		return
	}

	setAuthCookie(w, res.Token, res.Expires)
	h.rs.Send(ctx, w, http.StatusOK, res)
	log.InfoContext(ctx, "handler RefreshToken success", slog.String("user_id", res.UserID))
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "handler Logout start")

	if r.Method != http.MethodPost {
		log.WarnContext(ctx, "handler Logout wrong method")
		h.rs.Error(ctx, w, http.StatusMethodNotAllowed, "Logout", domain.ErrHTTPMethod, nil)
		return
	}

	clearAuthCookie(w)
	h.rs.Send(ctx, w, http.StatusOK, map[string]string{"message": "logged out"})
	log.InfoContext(ctx, "handler Logout success")
}

func rateLimitHandler(fn http.HandlerFunc) http.Handler {
	return middlewares.RateLimit(5, time.Minute)(fn)
}
