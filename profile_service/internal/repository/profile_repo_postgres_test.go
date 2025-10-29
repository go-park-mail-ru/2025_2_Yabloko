package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProfileRepoPostgresTestSuite struct {
	suite.Suite
	mock pgxmock.PgxPoolIface
	repo *ProfileRepoPostgres
}

func (s *ProfileRepoPostgresTestSuite) SetupTest() {
	mock, err := pgxmock.NewPool()
	require.NoError(s.T(), err)
	s.mock = mock
	s.repo = &ProfileRepoPostgres{
		db:  mock,
		log: logger.NewNilLogger(),
	}
}

func (s *ProfileRepoPostgresTestSuite) TearDownTest() {
	s.mock.Close()
}

func TestProfileRepoPostgresTestSuite(t *testing.T) {
	suite.Run(t, new(ProfileRepoPostgresTestSuite))
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_Success() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	expectedProfile := &domain.Profile{
		ID:        profileID,
		Email:     "test@example.com",
		Name:      stringPtr("John Doe"),
		Phone:     stringPtr("+123456789"),
		CityID:    stringPtr("city-123"),
		Address:   stringPtr("Test Address"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		expectedProfile.ID, expectedProfile.Email, expectedProfile.Name, expectedProfile.Phone,
		expectedProfile.CityID, expectedProfile.Address, expectedProfile.CreatedAt, expectedProfile.UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectedProfile, profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_SuccessWithNullFields() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	expectedProfile := &domain.Profile{
		ID:        profileID,
		Email:     "test@example.com",
		Name:      nil,
		Phone:     nil,
		CityID:    nil,
		Address:   nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		expectedProfile.ID, expectedProfile.Email, nil, nil, nil, nil, expectedProfile.CreatedAt, expectedProfile.UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectedProfile, profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_NotFound() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	s.mock.ExpectQuery(`SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnError(pgx.ErrNoRows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	assert.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_DatabaseError() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	dbError := errors.New("database error")

	s.mock.ExpectQuery(`SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnError(dbError)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfileByEmail_Success() {
	email := "test@example.com"
	expectedProfile := &domain.Profile{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        email,
		PasswordHash: "hashed_password",
		Name:         stringPtr("John Doe"),
		Phone:        stringPtr("+123456789"),
		CityID:       stringPtr("city-123"),
		Address:      stringPtr("Test Address"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "password_hash", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		expectedProfile.ID, expectedProfile.Email, expectedProfile.PasswordHash, expectedProfile.Name,
		expectedProfile.Phone, expectedProfile.CityID, expectedProfile.Address, expectedProfile.CreatedAt, expectedProfile.UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT id, email, password_hash, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfileByEmail(context.Background(), email)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectedProfile, profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfileByEmail_SuccessWithNullFields() {
	email := "test@example.com"
	expectedProfile := &domain.Profile{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        email,
		PasswordHash: "hashed_password",
		Name:         nil,
		Phone:        nil,
		CityID:       nil,
		Address:      nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "password_hash", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		expectedProfile.ID, expectedProfile.Email, expectedProfile.PasswordHash, nil, nil, nil, nil,
		expectedProfile.CreatedAt, expectedProfile.UpdatedAt,
	)

	s.mock.ExpectQuery(`SELECT id, email, password_hash, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \$1`).
		WithArgs(email).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfileByEmail(context.Background(), email)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectedProfile, profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfileByEmail_NotFound() {
	email := "notfound@example.com"

	s.mock.ExpectQuery(`SELECT id, email, password_hash, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(pgx.ErrNoRows)

	profile, err := s.repo.GetProfileByEmail(context.Background(), email)

	assert.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfileByEmail_DatabaseError() {
	email := "test@example.com"
	dbError := errors.New("database error")

	s.mock.ExpectQuery(`SELECT id, email, password_hash, name, phone, city_id, address, created_at, updated_at FROM account WHERE email = \$1`).
		WithArgs(email).
		WillReturnError(dbError)

	profile, err := s.repo.GetProfileByEmail(context.Background(), email)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_Success() {
	profile := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    stringPtr("Updated Name"),
		Phone:   stringPtr("+987654321"),
		CityID:  stringPtr("city-456"),
		Address: stringPtr("Updated Address"),
	}

	s.mock.ExpectExec(`UPDATE account SET name = \$1, phone = \$2, city_id = \$3, address = \$4 WHERE id = \$5`).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_SuccessWithNullFields() {
	profile := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    nil,
		Phone:   nil,
		CityID:  nil,
		Address: nil,
	}

	s.mock.ExpectExec(`UPDATE account SET name = \$1, phone = \$2, city_id = \$3, address = \$4 WHERE id = \$5`).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_SuccessPartialUpdate() {
	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Only Name Updated"),
	}

	s.mock.ExpectExec(`UPDATE account SET name = \$1, phone = \$2, city_id = \$3, address = \$4 WHERE id = \$5`).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_NotFound() {
	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Updated Name"),
	}

	s.mock.ExpectExec(`UPDATE account SET name = \$1, phone = \$2, city_id = \$3, address = \$4 WHERE id = \$5`).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_DatabaseError() {
	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Updated Name"),
	}
	dbError := errors.New("database error")

	s.mock.ExpectExec(`UPDATE account SET name = \$1, phone = \$2, city_id = \$3, address = \$4 WHERE id = \$5`).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).
		WillReturnError(dbError)

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestDeleteProfile_Success() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	s.mock.ExpectExec(`DELETE FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := s.repo.DeleteProfile(context.Background(), profileID)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestDeleteProfile_NotFound() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	s.mock.ExpectExec(`DELETE FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := s.repo.DeleteProfile(context.Background(), profileID)

	assert.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestDeleteProfile_DatabaseError() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	dbError := errors.New("database error")

	s.mock.ExpectExec(`DELETE FROM account WHERE id = \$1`).
		WithArgs(profileID).
		WillReturnError(dbError)

	err := s.repo.DeleteProfile(context.Background(), profileID)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestCreateProfile_Success() {
	profile := &domain.Profile{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "newuser@example.com",
		PasswordHash: "hashedpassword123",
	}

	s.mock.ExpectExec(`INSERT INTO account \(id, email, password_hash\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(profile.ID, profile.Email, profile.PasswordHash).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := s.repo.CreateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestCreateProfile_DuplicateEmail() {
	profile := &domain.Profile{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword123",
	}
	dbError := errors.New("duplicate key value")

	s.mock.ExpectExec(`INSERT INTO account \(id, email, password_hash\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(profile.ID, profile.Email, profile.PasswordHash).
		WillReturnError(dbError)

	err := s.repo.CreateProfile(context.Background(), profile)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "duplicate key")
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestCreateProfile_DatabaseError() {
	profile := &domain.Profile{
		ID:           "550e8400-e29b-41d4-a716-446655440000",
		Email:        "newuser@example.com",
		PasswordHash: "hashedpassword123",
	}
	dbError := errors.New("database error")

	s.mock.ExpectExec(`INSERT INTO account \(id, email, password_hash\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(profile.ID, profile.Email, profile.PasswordHash).
		WillReturnError(dbError)

	err := s.repo.CreateProfile(context.Background(), profile)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestAllMethods_ContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	profileID := "550e8400-e29b-41d4-a716-446655440000"
	profile := &domain.Profile{
		ID:    profileID,
		Email: "test@example.com",
	}

	s.mock.ExpectQuery(`SELECT.*`).WithArgs(profileID).WillReturnError(context.Canceled)
	s.mock.ExpectExec(`UPDATE.*`).WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.ID).WillReturnError(context.Canceled)
	s.mock.ExpectExec(`DELETE.*`).WithArgs(profileID).WillReturnError(context.Canceled)
	s.mock.ExpectExec(`INSERT.*`).WithArgs(profile.ID, profile.Email, profile.PasswordHash).WillReturnError(context.Canceled)

	_, err := s.repo.GetProfile(ctx, profileID)
	assert.ErrorIs(s.T(), err, context.Canceled)

	err = s.repo.UpdateProfile(ctx, profile)
	assert.ErrorIs(s.T(), err, context.Canceled)

	err = s.repo.DeleteProfile(ctx, profileID)
	assert.ErrorIs(s.T(), err, context.Canceled)

	err = s.repo.CreateProfile(ctx, profile)
	assert.ErrorIs(s.T(), err, context.Canceled)

	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func stringPtr(s string) *string {
	return &s
}
