package usecase

import (
	"apple_backend/auth_service/internal/delivery/transport"
	"apple_backend/auth_service/internal/domain"
	"apple_backend/pkg/blacklist"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, email, hashedPassword string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UserExists(ctx context.Context, email string) (bool, error)
}

type authUseCase struct {
	repo      AuthRepository
	blacklist blacklist.TokenBlacklist
	jwtSecret []byte
	logger    *slog.Logger
}

func NewAuthUseCase(
	repo AuthRepository,
	tb blacklist.TokenBlacklist,
	secretKey string,
	logger *slog.Logger,
) *authUseCase {
	return &authUseCase{
		repo:      repo,
		blacklist: tb,
		jwtSecret: []byte(secretKey),
		logger:    logger,
	}
}

func (uc *authUseCase) Register(ctx context.Context, email, password string) (*transport.AuthResult, error) {
	if err := uc.validateRegistrationInput(email, password); err != nil {
		return nil, err
	}
	exists, err := uc.repo.UserExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}
	hashed, err := uc.hashPassword(password)
	if err != nil {
		return nil, err
	}
	user, err := uc.repo.CreateUser(ctx, email, hashed)
	if err != nil {
		return nil, err
	}
	token, err := uc.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &transport.AuthResult{
		UserID:  user.ID,
		Email:   user.Email,
		Token:   token,
		Expires: time.Now().Add(24 * time.Hour),
	}, nil
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (*transport.AuthResult, error) {
	if err := uc.validateLoginInput(email, password); err != nil {
		return nil, err
	}
	user, err := uc.repo.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if !uc.checkPasswordHash(password, user.PasswordHash) {
		return nil, domain.ErrInvalidPassword
	}
	token, err := uc.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &transport.AuthResult{
		UserID:  user.ID,
		Email:   user.Email,
		Token:   token,
		Expires: time.Now().Add(24 * time.Hour),
	}, nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, tokenString string) (*transport.AuthResult, error) {
	isBlacklisted, err := uc.blacklist.IsBlacklisted(ctx, tokenString)
	if err != nil {
		uc.logger.ErrorContext(ctx, "blacklist check failed", slog.Any("err", err))
		return nil, err
	}
	if isBlacklisted {
		uc.logger.WarnContext(ctx, "attempt to refresh blacklisted token")
		return nil, domain.ErrTokenBlacklisted
	}

	claims, err := uc.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	user, err := uc.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	newTok, err := uc.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	return &transport.AuthResult{
		UserID:  user.ID,
		Email:   user.Email,
		Token:   newTok,
		Expires: time.Now().Add(24 * time.Hour),
	}, nil
}

func (uc *authUseCase) VerifyToken(ctx context.Context, tokenString string) (*transport.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &transport.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return uc.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*transport.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, domain.ErrInvalidToken
}

func (uc *authUseCase) LogoutToken(ctx context.Context, tokenString string) error {
	claims, err := uc.VerifyToken(ctx, tokenString)
	if err != nil {
		uc.logger.WarnContext(ctx, "logout with invalid token", slog.Any("err", err))
		return nil
	}

	expiresAt := time.Unix(claims.ExpiresAt.Unix(), 0)
	err = uc.blacklist.InsertToBlacklist(ctx, tokenString, expiresAt)
	if err != nil {
		uc.logger.ErrorContext(ctx, "failed to blacklist token", slog.Any("err", err))
		return err
	}

	uc.logger.InfoContext(ctx, "token blacklisted on logout", slog.String("user_id", claims.UserID))
	return nil
}

func (uc *authUseCase) ValidateEmail(ctx context.Context, email string) error {
	return uc.validateEmailFormat(email)
}

func (uc *authUseCase) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return uc.repo.GetUserByID(ctx, userID)
}

func (uc *authUseCase) validateRegistrationInput(email, password string) error {
	if err := uc.validateEmailFormat(email); err != nil {
		return err
	}
	return uc.validatePasswordSecurity(email, password)
}

func (uc *authUseCase) validateLoginInput(email, password string) error {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" {
		return domain.ErrInvalidEmail
	}
	if password == "" {
		return domain.ErrInvalidPassword
	}
	return uc.validateEmailFormat(email)
}

func (uc *authUseCase) validateEmailFormat(email string) error {
	email = strings.TrimSpace(email)
	if email == "" || len(email) < 3 || len(email) > 254 {
		return domain.ErrInvalidEmail
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return domain.ErrInvalidEmail
	}
	return nil
}

func (uc *authUseCase) validatePasswordSecurity(email, password string) error {
	if len(password) < 8 {
		return domain.ErrWeakPassword
	}
	var up, low, num, spec bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			up = true
		case unicode.IsLower(r):
			low = true
		case unicode.IsNumber(r):
			num = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			spec = true
		}
	}
	if !up || !low || !num || !spec {
		return domain.ErrWeakPassword
	}
	if strings.EqualFold(email, password) {
		return domain.ErrWeakPassword
	}
	return nil
}

func (uc *authUseCase) hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (uc *authUseCase) checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (uc *authUseCase) generateToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(uc.jwtSecret)
}
