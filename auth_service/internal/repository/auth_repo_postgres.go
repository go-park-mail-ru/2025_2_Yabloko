package repository

import (
	"apple_backend/auth_service/internal/domain"
	"context"
	_ "embed"

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
	id := uuid.NewString()

	var u domain.User
	err := r.db.QueryRow(ctx, createUserSQL, id, email, hashedPassword).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, err
	}
	return &u, nil
}

//go:embed sql/auth/get_user_by_email.sql
var getUserByEmailSQL string

func (r *AuthRepoPostgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, getUserByEmailSQL, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	return &u, err
}

//go:embed sql/auth/get_user_by_id.sql
var getUserByIDSQL string

func (r *AuthRepoPostgres) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx, getUserByIDSQL, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	return &u, err
}

//go:embed sql/auth/user_exists.sql
var userExistsSQL string

func (r *AuthRepoPostgres) UserExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, userExistsSQL, email).Scan(&exists)
	return exists, err
}
