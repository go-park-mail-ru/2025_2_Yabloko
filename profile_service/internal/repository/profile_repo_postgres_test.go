package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestProfileRepoPostgres_GetProfile(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	log := logger.NewNilLogger()
	repo := &ProfileRepoPostgres{
		db:  mock,
		log: log,
	}

	t.Run("Успешное получение профиля", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"
		expectedEmail := "test@example.com"
		expectedName := "John Doe"
		expectedPhone := "+123456789"
		expectedCityID := "city-123"
		expectedAddress := "Test Address"
		now := time.Now()

		rows := mock.NewRows([]string{
			"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
		}).AddRow(
			expectedID, expectedEmail, &expectedName, &expectedPhone, &expectedCityID,
			&expectedAddress, now, now,
		)

		mock.ExpectQuery("SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(context.Background(), expectedID)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.Equal(t, expectedID, profile.ID)
		require.Equal(t, expectedEmail, profile.Email)
		require.Equal(t, &expectedName, profile.Name)
		require.Equal(t, &expectedPhone, profile.Phone)
		require.Equal(t, &expectedCityID, profile.CityID)
		require.Equal(t, &expectedAddress, profile.Address)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Профиль не найден", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"

		mock.ExpectQuery("SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnError(domain.ErrProfileNotFound)

		profile, err := repo.GetProfile(context.Background(), expectedID)
		require.Error(t, err)
		require.Nil(t, profile)
		require.Equal(t, domain.ErrProfileNotFound, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка БД при получении профиля", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"

		mock.ExpectQuery("SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnError(errors.New("database error"))

		profile, err := repo.GetProfile(context.Background(), expectedID)
		require.Error(t, err)
		require.Nil(t, profile)
		require.Contains(t, err.Error(), "database error")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProfileRepoPostgres_GetProfileByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	log := logger.NewNilLogger()
	repo := &ProfileRepoPostgres{
		db:  mock,
		log: log,
	}

	t.Run("Успешное получение профиля по email", func(t *testing.T) {
		expectedEmail := "test@example.com"
		expectedID := "550e8400-e29b-41d4-a716-446655440000"
		expectedName := "John Doe"
		expectedPhone := "+123456789"
		expectedCityID := "city-123"
		expectedAddress := "Test Address"
		now := time.Now()

		rows := mock.NewRows([]string{
			"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
		}).AddRow(
			expectedID, expectedEmail, &expectedName, &expectedPhone, &expectedCityID,
			&expectedAddress, now, now,
		)

		mock.ExpectQuery("SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \\$1").
			WithArgs(expectedEmail).
			WillReturnRows(rows)

		profile, err := repo.GetProfileByEmail(context.Background(), expectedEmail)
		require.NoError(t, err)
		require.NotNil(t, profile)
		require.Equal(t, expectedEmail, profile.Email)
		require.Equal(t, expectedID, profile.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Профиль не найден по email", func(t *testing.T) {
		expectedEmail := "notfound@example.com"

		mock.ExpectQuery("SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \\$1").
			WithArgs(expectedEmail).
			WillReturnError(domain.ErrProfileNotFound)

		profile, err := repo.GetProfileByEmail(context.Background(), expectedEmail)
		require.Error(t, err)
		require.Nil(t, profile)
		require.Equal(t, domain.ErrProfileNotFound, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProfileRepoPostgres_UpdateProfile(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	log := logger.NewNilLogger()
	repo := &ProfileRepoPostgres{
		db:  mock,
		log: log,
	}

	t.Run("Успешное обновление профиля", func(t *testing.T) {
		profile := &domain.Profile{
			ID:      "550e8400-e29b-41d4-a716-446655440000",
			Name:    stringPtr("Updated Name"),
			Phone:   stringPtr("+987654321"),
			CityID:  stringPtr("city-456"),
			Address: stringPtr("Updated Address"),
		}

		mock.ExpectExec("UPDATE account SET name = \\$1, phone = \\$2, city_id = \\$3, address = \\$4 WHERE id = \\$5").
			WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := repo.UpdateProfile(context.Background(), profile)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Профиль не найден при обновлении", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: stringPtr("Updated Name"),
		}

		mock.ExpectExec("UPDATE account SET name = \\$1, phone = \\$2, city_id = \\$3, address = \\$4 WHERE id = \\$5").
			WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err := repo.UpdateProfile(context.Background(), profile)
		require.Error(t, err)
		require.Equal(t, domain.ErrProfileNotFound, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка БД при обновлении", func(t *testing.T) {
		profile := &domain.Profile{
			ID:   "550e8400-e29b-41d4-a716-446655440000",
			Name: stringPtr("Updated Name"),
		}

		mock.ExpectExec("UPDATE account SET name = \\$1, phone = \\$2, city_id = \\$3, address = \\$4 WHERE id = \\$5").
			WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
			WillReturnError(errors.New("database error"))

		err := repo.UpdateProfile(context.Background(), profile)
		require.Error(t, err)
		require.Contains(t, err.Error(), "database error")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProfileRepoPostgres_DeleteProfile(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	log := logger.NewNilLogger()
	repo := &ProfileRepoPostgres{
		db:  mock,
		log: log,
	}

	t.Run("Успешное удаление профиля", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"

		mock.ExpectExec("DELETE FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err := repo.DeleteProfile(context.Background(), expectedID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Профиль не найден при удалении", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"

		mock.ExpectExec("DELETE FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err := repo.DeleteProfile(context.Background(), expectedID)
		require.Error(t, err)
		require.Equal(t, domain.ErrProfileNotFound, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка БД при удалении", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"

		mock.ExpectExec("DELETE FROM account WHERE id = \\$1").
			WithArgs(expectedID).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteProfile(context.Background(), expectedID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "db error")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProfileRepoPostgres_CreateProfile(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	log := logger.NewNilLogger()
	repo := &ProfileRepoPostgres{
		db:  mock,
		log: log,
	}

	t.Run("Успешное создание профиля", func(t *testing.T) {
		profile := &domain.Profile{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			Email:        "newuser@example.com",
			PasswordHash: "hashedpassword123",
		}

		mock.ExpectExec("INSERT INTO account \\(id, email, password_hash, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\\)").
			WithArgs(profile.ID, profile.Email, profile.PasswordHash).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.CreateProfile(context.Background(), profile)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Ошибка дублирования email", func(t *testing.T) {
		profile := &domain.Profile{
			ID:           "550e8400-e29b-41d4-a716-446655440000",
			Email:        "existing@example.com",
			PasswordHash: "hashedpassword123",
		}

		mock.ExpectExec("INSERT INTO account \\(id, email, password_hash, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP\\)").
			WithArgs(profile.ID, profile.Email, profile.PasswordHash).
			WillReturnError(errors.New("duplicate key value"))

		err := repo.CreateProfile(context.Background(), profile)
		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate key")
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func stringPtr(s string) *string {
	return &s
}
