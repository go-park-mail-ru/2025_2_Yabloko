package usecase

import (
	"apple_backend/auth_service/internal/domain"
	mocks "apple_backend/auth_service/internal/usecase/mock"
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAuthRepository(ctrl)
	uc := NewAuthUseCase(repo, "test-secret")

	repo.EXPECT().UserExists(gomock.Any(), "u@ex.com").Return(false, nil)
	repo.EXPECT().
		CreateUser(gomock.Any(), "u@ex.com", gomock.Any()).
		DoAndReturn(func(_ context.Context, email, hash string) (*domain.User, error) {
			return &domain.User{ID: "u1", Email: email, PasswordHash: hash, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		})

	res, err := uc.Register(context.Background(), "u@ex.com", "Str0ng!Pass")
	if err != nil || res.Token == "" || res.UserID == "" {
		t.Fatalf("register failed: err=%v res=%+v", err, res)
	}
}

func TestLogin_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAuthRepository(ctrl)
	uc := NewAuthUseCase(repo, "test-secret")

	hash, _ := bcrypt.GenerateFromPassword([]byte("Str0ng!Pass"), bcrypt.DefaultCost)
	user := &domain.User{ID: "u1", Email: "u@ex.com", PasswordHash: string(hash)}

	repo.EXPECT().GetUserByEmail(gomock.Any(), "u@ex.com").Return(user, nil)

	res, err := uc.Login(context.Background(), "u@ex.com", "Str0ng!Pass")
	if err != nil || res.Token == "" {
		t.Fatalf("login failed: err=%v", err)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAuthRepository(ctrl)
	uc := NewAuthUseCase(repo, "test-secret")

	if _, err := uc.RefreshToken(context.Background(), "bad.token"); err == nil {
		t.Fatal("expected error for bad token")
	}
}
