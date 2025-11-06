package http

import (
	"apple_backend/auth_service/internal/delivery/transport"
	"apple_backend/auth_service/internal/domain"
	"apple_backend/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mock "apple_backend/auth_service/internal/delivery/http/mock"
	"log/slog"

	"github.com/golang/mock/gomock"
)

type fakeUC struct{}

func (f *fakeUC) Register(_ interface{}, _ ...interface{}) {}
func (f *fakeUC) Login(_ interface{}, _ ...interface{})    {}

type stubUC struct{}

func (s *stubUC) Register(_ interface{}, _ ...interface{}) {}

type testUC struct{}

func (tuc *testUC) Register(_ interface{}, _ ...interface{}) {}

type ucStub struct{}

func (u *ucStub) Register(_ interface{}, _ ...interface{}) {}

type ucFake struct{}

func (uc *ucFake) Register(_ interface{}, _ ...interface{}) {}

type authUC struct{}

func (a *authUC) Register(_ interface{}, _ ...interface{}) {}

type okUC struct{}

func (ok *okUC) Register(_ interface{}, _ ...interface{}) {}

type hUC struct{}

func (h *hUC) Register(_ interface{}, _ ...interface{}) {}

type handlerUC struct{}

func (handlerUC) Register(_ context.Context, email, password string) (*transport.AuthResult, error) {
	return &transport.AuthResult{
		UserID:  "u1",
		Email:   email,
		Token:   "tok",
		Expires: time.Now().Add(time.Hour),
	}, nil
}
func (handlerUC) Login(_ context.Context, email, password string) (*transport.AuthResult, error) {
	return &transport.AuthResult{
		UserID:  "u1",
		Email:   email,
		Token:   "tok",
		Expires: time.Now().Add(time.Hour),
	}, nil
}
func (handlerUC) RefreshToken(_ context.Context, token string) (*transport.AuthResult, error) {
	return &transport.AuthResult{
		UserID:  "u1",
		Email:   "u@ex.com",
		Token:   "tok2",
		Expires: time.Now().Add(time.Hour),
	}, nil
}
func (handlerUC) VerifyToken(_ context.Context, token string) (*transport.Claims, error) {
	return &transport.Claims{UserID: "u1", Email: "u@ex.com"}, nil
}
func (handlerUC) ValidateEmail(_ context.Context, email string) error { return nil }
func (handlerUC) GetUserByID(_ context.Context, id string) (*domain.User, error) {
	return &domain.User{ID: id, Email: "u@ex.com"}, nil
}

var _ AuthUseCaseInterface = handlerUC{}

func TestAuthHandler_Register(t *testing.T) {
	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)
	h := NewAuthHandler(handlerUC{}, logg)

	body, _ := json.Marshal(transport.RegisterRequest{Email: "u@ex.com", Password: "Str0ng!Pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var res transport.AuthResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if res.Token == "" || res.UserID == "" {
		t.Fatal("empty token or user id")
	}
}

func TestAuthHandler_Login(t *testing.T) {
	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)
	h := NewAuthHandler(handlerUC{}, logg)

	body, _ := json.Marshal(transport.LoginRequest{Email: "u@ex.com", Password: "Str0ng!Pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)
	h := NewAuthHandler(handlerUC{}, logg)

	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "tok"})
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegister_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)

	uc := mock.NewMockAuthUseCaseInterface(ctrl)

	h := NewAuthHandler(uc, logg)

	reqBody, _ := json.Marshal(transport.RegisterRequest{Email: "u@ex.com", Password: "Str0ng!Pass"})
	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/signup", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	uc.EXPECT().
		Register(gomock.Any(), "u@ex.com", "Str0ng!Pass").
		Return(&transport.AuthResult{
			UserID: "u1", Email: "u@ex.com", Token: "tok", Expires: time.Now().Add(time.Hour),
		}, nil)

	h.Register(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestLogin_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)
	uc := mock.NewMockAuthUseCaseInterface(ctrl)
	h := NewAuthHandler(uc, logg)

	reqBody, _ := json.Marshal(transport.LoginRequest{Email: "u@ex.com", Password: "bad"})
	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	uc.EXPECT().
		Login(gomock.Any(), "u@ex.com", "bad").
		Return(nil, domain.ErrInvalidPassword)

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestRefresh_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logg := logger.NewLogger("auth_handler_test", slog.LevelDebug)
	uc := mock.NewMockAuthUseCaseInterface(ctrl)
	h := NewAuthHandler(uc, logg)

	req := httptest.NewRequest(http.MethodPost, "/api/v0/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "tok"})
	w := httptest.NewRecorder()

	uc.EXPECT().
		RefreshToken(gomock.Any(), "tok").
		Return(&transport.AuthResult{UserID: "u1", Email: "u@ex.com", Token: "tok2", Expires: time.Now().Add(time.Hour)}, nil)

	h.RefreshToken(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
}
