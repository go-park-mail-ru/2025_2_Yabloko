package repository

import (
	"apple_backend/auth_service/internal/domain"
	"apple_backend/pkg/logger"
	"context"
	_ "embed"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepoPostgres struct {
	db *pgxpool.Pool
}

func NewAuthRepoPostgres(db *pgxpool.Pool) *AuthRepoPostgres {
	return &AuthRepoPostgres{db: db}
}

//go:embed sql/auth/create_user.sql
var createUserSQL string

func (r *AuthRepoPostgres) CreateUser(ctx context.Context, email, hashedPassword string) (*domain.User, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo CreateUser start", slog.String("email", email))

	id := uuid.NewString()

	var u domain.User
	err := r.db.QueryRow(ctx, createUserSQL, id, email, hashedPassword).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			log.WarnContext(ctx, "repo CreateUser user already exists", slog.String("email", email))
			return nil, domain.ErrUserAlreadyExists
		}
		log.ErrorContext(ctx, "repo CreateUser database error", slog.Any("err", err), slog.String("email", email))
		return nil, err
	}

	log.InfoContext(ctx, "repo CreateUser success", slog.String("user_id", u.ID), slog.String("email", email))
	return &u, nil
}

//go:embed sql/auth/get_user_by_email.sql
var getUserByEmailSQL string

func (r *AuthRepoPostgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo GetUserByEmail start", slog.String("email", email))

	var u domain.User
	err := r.db.QueryRow(ctx, getUserByEmailSQL, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		log.WarnContext(ctx, "repo GetUserByEmail user not found", slog.String("email", email))
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "repo GetUserByEmail database error", slog.Any("err", err), slog.String("email", email))
		return nil, err
	}

	log.InfoContext(ctx, "repo GetUserByEmail success", slog.String("user_id", u.ID), slog.String("email", email))
	return &u, nil
}

//go:embed sql/auth/get_user_by_id.sql
var getUserByIDSQL string

func (r *AuthRepoPostgres) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "repo GetUserByID start", slog.String("user_id", id))

	var u domain.User
	err := r.db.QueryRow(ctx, getUserByIDSQL, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		log.WarnContext(ctx, "repo GetUserByID user not found", slog.String("user_id", id))
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		log.ErrorContext(ctx, "repo GetUserByID database error", slog.Any("err", err), slog.String("user_id", id))
		return nil, err
	}

	log.InfoContext(ctx, "repo GetUserByID success", slog.String("user_id", u.ID))
	return &u, nil
}

//go:embed sql/auth/user_exists.sql
var userExistsSQL string

func (r *AuthRepoPostgres) UserExists(ctx context.Context, email string) (bool, error) {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "repo UserExists check", slog.String("email", email))

	var exists bool
	err := r.db.QueryRow(ctx, userExistsSQL, email).Scan(&exists)
	if err != nil {
		log.ErrorContext(ctx, "repo UserExists database error", slog.Any("err", err), slog.String("email", email))
		return false, err
	}

	log.DebugContext(ctx, "repo UserExists result", slog.Bool("exists", exists), slog.String("email", email))
	return exists, nil
}
