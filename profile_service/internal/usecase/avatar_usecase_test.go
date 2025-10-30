package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"apple_backend/profile_service/internal/domain"
	"apple_backend/profile_service/internal/usecase/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) { return 0, errors.New("read error") }

type seekerReader struct {
	*bytes.Reader
}

func TestAvatarUsecase_FullCoverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmpDir := t.TempDir()
	mockRepo := mock.NewMockProfileRepository(ctrl)
	uc := NewAvatarUsecase(mockRepo, "http://localhost/", tmpDir)

	validImage := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	userID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("NewAvatarUsecase", func(t *testing.T) {
		obj := NewAvatarUsecase(mockRepo, "http://localhost/", tmpDir)
		require.NotNil(t, obj)
		require.Equal(t, tmpDir, obj.uploadPath)
	})

	tests := []struct {
		name        string
		userID      string
		reader      io.Reader
		setupMock   func()
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "Невалидный UUID",
			userID:      "invalid-uuid",
			reader:      bytes.NewReader(validImage),
			setupMock:   func() {},
			expectError: true,
			errorCheck:  func(err error) bool { return errors.Is(err, domain.ErrInvalidProfileData) },
		},
		{
			name:        "Пустой файл",
			userID:      userID,
			reader:      bytes.NewReader([]byte{}),
			setupMock:   func() {},
			expectError: true,
			errorCheck:  func(err error) bool { return errors.Is(err, domain.ErrInvalidFileType) },
		},
		{
			name:        "Невалидный MIME",
			userID:      userID,
			reader:      bytes.NewReader([]byte("plain text")),
			setupMock:   func() {},
			expectError: true,
			errorCheck:  func(err error) bool { return errors.Is(err, domain.ErrInvalidFileType) },
		},
		{
			name:        "Ошибка io.ReadFull",
			userID:      userID,
			reader:      &failingReader{},
			setupMock:   func() {},
			expectError: true,
		},
		{
			name:   "Reader с io.Seeker",
			userID: userID,
			reader: &seekerReader{bytes.NewReader(validImage)},
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), userID).Return(&domain.Profile{ID: userID}, nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "Reader без io.Seeker (MultiReader)",
			userID: userID,
			reader: io.LimitReader(bytes.NewReader(validImage), int64(len(validImage))),
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), userID).Return(&domain.Profile{ID: userID}, nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "Ошибка GetProfile",
			userID: userID,
			reader: bytes.NewReader(validImage),
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), userID).Return(nil, errors.New("get error"))
			},
			expectError: true,
			errorCheck:  func(err error) bool { return err.Error() == "get error" },
		},
		{
			name:   "Ошибка UpdateProfile",
			userID: userID,
			reader: bytes.NewReader(validImage),
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), userID).Return(&domain.Profile{ID: userID}, nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(errors.New("update error"))
			},
			expectError: true,
			errorCheck:  func(err error) bool { return err.Error() == "update error" },
		},
		{
			name:   "Замена старого аватара с битым URL",
			userID: userID,
			reader: bytes.NewReader(validImage),
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(gomock.Any(), userID).Return(&domain.Profile{
					ID:        userID,
					AvatarURL: func() *string { s := "://invalid-url"; return &s }(),
				}, nil)
				mockRepo.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			url, err := uc.UploadAvatar(context.Background(), tt.userID, tt.reader, &multipart.FileHeader{})

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					require.True(t, tt.errorCheck(err))
				}
				require.Empty(t, url)
			} else {
				require.NoError(t, err)
				require.Contains(t, url, tmpDir)

				filename := filepath.Base(url)
				_, ferr := os.Stat(filepath.Join(tmpDir, filename))
				require.NoError(t, ferr)
			}
		})
	}
}
