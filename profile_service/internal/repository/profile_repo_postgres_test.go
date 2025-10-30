package repository

import (
	"apple_backend/pkg/logger"
	"apple_backend/profile_service/internal/domain"
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
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

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	require.NoError(s.T(), err)
	assert.Equal(s.T(), expectedProfile, profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_NotFound() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnError(pgx.ErrNoRows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	assert.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_ContextDeadlineExceeded() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	profileID := "550e8400-e29b-41d4-a716-446655440000"
	s.mock.ExpectQuery(`SELECT.*`).WithArgs(profileID).WillReturnError(context.DeadlineExceeded)

	_, err := s.repo.GetProfile(ctx, profileID)
	require.ErrorIs(s.T(), err, context.DeadlineExceeded)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_DatabaseError() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	dbError := errors.New("database error")

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnError(dbError)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "database error")
	assert.Nil(s.T(), profile)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_ScanError() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		profileID, "test@example.com", "John", "+123456789", "city-1", "Addr 1",
		"not-a-time", "not-a-time",
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)

	require.Error(s.T(), err)
	require.Nil(s.T(), profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_Success() {
	profile := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    stringPtr("Updated Name"),
		Phone:   stringPtr("+987654321"),
		CityID:  stringPtr("city-456"),
		Address: stringPtr("Updated Address"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
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

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_ContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	profile := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    stringPtr("Name"),
		Phone:   stringPtr("123"),
		CityID:  stringPtr("city"),
		Address: stringPtr("addr"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnError(context.Canceled)

	err := s.repo.UpdateProfile(ctx, profile)
	require.ErrorIs(s.T(), err, context.Canceled)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_SuccessPartialUpdate() {
	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Only Name Updated"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)

	assert.NoError(s.T(), err)
	assert.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_WithAvatarURL() {
	profile := &domain.Profile{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:      stringPtr("Name"),
		Phone:     stringPtr("+123456789"),
		CityID:    stringPtr("city-001"),
		Address:   stringPtr("Address 1"),
		AvatarURL: stringPtr("http://example.com/avatar.jpg"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_WithEmptyAvatarURL() {
	empty := ""
	profile := &domain.Profile{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:      stringPtr("Name"),
		Phone:     stringPtr("+123456789"),
		CityID:    stringPtr("city-001"),
		Address:   stringPtr("Address 1"),
		AvatarURL: &empty,
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), profile)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_NotFound() {
	profile := &domain.Profile{
		ID:   "550e8400-e29b-41d4-a716-446655440000",
		Name: stringPtr("Updated Name"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
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

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
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

func (s *ProfileRepoPostgresTestSuite) TestDeleteProfile_ContextCanceled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	profileID := "550e8400-e29b-41d4-a716-446655440000"
	s.mock.ExpectExec(`DELETE FROM account WHERE id = \$1`).WithArgs(profileID).WillReturnError(context.Canceled)

	err := s.repo.DeleteProfile(ctx, profileID)
	require.ErrorIs(s.T(), err, context.Canceled)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
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

	pgErr := &pgconn.PgError{
		Code:    "23505",
		Message: "duplicate key value violates unique constraint \"account_email_key\"",
	}

	s.mock.ExpectExec(`INSERT INTO account \(id, email, password_hash\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(profile.ID, profile.Email, profile.PasswordHash).
		WillReturnError(pgErr)

	err := s.repo.CreateProfile(context.Background(), profile)

	assert.ErrorIs(s.T(), err, domain.ErrProfileExist)
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

func (s *ProfileRepoPostgresTestSuite) TestCreateProfile_PgErrorOtherThan23505() {
	profile := &domain.Profile{
		ID:           "id-001",
		Email:        "user@example.com",
		PasswordHash: "hash123",
	}
	pgErr := &pgconn.PgError{
		Code:    "12345",
		Message: "some other pg error",
	}

	s.mock.ExpectExec(`INSERT INTO account \(id, email, password_hash\) VALUES \(\$1, \$2, \$3\)`).
		WithArgs(profile.ID, profile.Email, profile.PasswordHash).
		WillReturnError(pgErr)

	err := s.repo.CreateProfile(context.Background(), profile)
	require.Error(s.T(), err)
	require.Equal(s.T(), pgErr, err)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
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
	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(profile.Name, profile.Phone, profile.CityID, profile.Address, profile.AvatarURL, profile.ID).
		WillReturnError(context.Canceled)
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

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_PartialNullMix() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"
	created := time.Now()
	updated := time.Now()

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		profileID,
		"mix@example.com",
		"John Mix",
		nil,
		"city-mix",
		nil,
		created,
		updated,
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.NoError(s.T(), err)

	require.Equal(s.T(), profileID, profile.ID)
	require.Equal(s.T(), "mix@example.com", profile.Email)
	require.NotNil(s.T(), profile.Name)
	require.Equal(s.T(), "John Mix", *profile.Name)
	require.Nil(s.T(), profile.Phone)
	require.NotNil(s.T(), profile.CityID)
	require.Equal(s.T(), "city-mix", *profile.CityID)
	require.Nil(s.T(), profile.Address)

	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_NoRowsResult() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	})
	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.ErrorIs(s.T(), err, domain.ErrProfileNotFound)
	require.Nil(s.T(), profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_RowError() {
	profileID := "550e8400-e29b-41d4-a716-446655440000"

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "created_at", "updated_at",
	}).AddRow(
		profileID, "test@example.com", "John", nil, nil, nil, time.Now(), time.Now(),
	)
	rows.RowError(0, errors.New("row error"))

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.Error(s.T(), err)
	require.Nil(s.T(), profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func stringPtr(s string) *string {
	return &s
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_OnlyAvatarURL() {
	avatar := "http://example.com/a.jpg"
	p := &domain.Profile{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		AvatarURL: &avatar,
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(p.Name, p.Phone, p.CityID, p.Address, p.AvatarURL, p.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := s.repo.UpdateProfile(context.Background(), p)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestUpdateProfile_DeadlineExceeded() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &domain.Profile{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    stringPtr("n"),
		Phone:   stringPtr("p"),
		CityID:  stringPtr("c"),
		Address: stringPtr("a"),
	}

	s.mock.ExpectExec(regexp.QuoteMeta(updateProfileSQL)).
		WithArgs(p.Name, p.Phone, p.CityID, p.Address, p.AvatarURL, p.ID).
		WillReturnError(context.DeadlineExceeded)

	err := s.repo.UpdateProfile(ctx, p)
	require.ErrorIs(s.T(), err, context.DeadlineExceeded)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_WithAvatarURL() {
	profileID := "550e8400-e29b-41d4-a716-446655440001"
	avatar := "http://example.com/avatar.png"
	expected := &domain.Profile{
		ID:        profileID,
		Email:     "avatar@example.com",
		Name:      stringPtr("Avatar Name"),
		Phone:     stringPtr("+111222333"),
		CityID:    stringPtr("city-avatar"),
		Address:   stringPtr("Some Address"),
		AvatarURL: &avatar,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "avatar_url", "created_at", "updated_at",
	}).AddRow(
		expected.ID, expected.Email, expected.Name, expected.Phone, expected.CityID,
		expected.Address, expected.AvatarURL, expected.CreatedAt, expected.UpdatedAt,
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), expected, profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_AllNullFields() {
	profileID := "550e8400-e29b-41d4-a716-446655440002"
	expected := &domain.Profile{
		ID:        profileID,
		Email:     "nullfields@example.com",
		Name:      nil,
		Phone:     nil,
		CityID:    nil,
		Address:   nil,
		AvatarURL: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "avatar_url", "created_at", "updated_at",
	}).AddRow(
		expected.ID, expected.Email, nil, nil, nil, nil, nil, expected.CreatedAt, expected.UpdatedAt,
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), expected, profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_ScanEmailNil() {
	profileID := "550e8400-e29b-41d4-a716-446655440003"

	rows := s.mock.NewRows([]string{
		"id", "email", "name", "phone", "city_id", "address", "avatar_url", "created_at", "updated_at",
	}).AddRow(
		profileID, nil, "John", "+123", "city-1", "Addr", nil, time.Now(), time.Now(),
	)

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnRows(rows)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.Error(s.T(), err)
	require.Nil(s.T(), profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

func (s *ProfileRepoPostgresTestSuite) TestGetProfile_QueryError() {
	profileID := "550e8400-e29b-41d4-a716-446655440004"
	dbError := errors.New("unexpected db error")

	s.mock.ExpectQuery(regexp.QuoteMeta(selectProfileSQL)).
		WithArgs(profileID).
		WillReturnError(dbError)

	profile, err := s.repo.GetProfile(context.Background(), profileID)
	require.ErrorIs(s.T(), err, dbError)
	require.Nil(s.T(), profile)
	require.NoError(s.T(), s.mock.ExpectationsWereMet())
}

const (
	selectProfileSQL = `SELECT id, email, name, phone, city_id, address, created_at, updated_at FROM account WHERE id = $1`
	updateProfileSQL = `
		UPDATE account
		SET name = $1,
			phone = $2,
			city_id = $3,
			address = $4,
			avatar_url = $5
		WHERE id = $6
	`
)
