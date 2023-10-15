package service_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	mock_repository "github.com/b0shka/backend/internal/repository/postgresql/mocks"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/service"
	mock_worker "github.com/b0shka/backend/internal/worker/mocks"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var ErrInternalServerError = errors.New("test: internal server error")

func mockAuthService(t *testing.T) (*service.AuthService, *mock_repository.MockStore) {
	repoCtl := gomock.NewController(t)
	defer repoCtl.Finish()

	workerCtl := gomock.NewController(t)
	defer workerCtl.Finish()

	repo := mock_repository.NewMockStore(repoCtl)
	worker := mock_worker.NewMockTaskDistributor(workerCtl)
	authService := service.NewAuthService(
		repo,
		&hash.SHA256Hasher{},
		&auth.JWTManager{},
		&otp.TOTPGenerator{},
		config.AuthConfig{},
		worker,
	)

	return authService, repo
}

// func TestUsersService_SendCodeEmail(t *testing.T) {
// 	authService, userRepo := mockAuthService(t)

// 	ctx := context.Background()
// 	userRepo.EXPECT().CreateVerifyEmail(ctx, gomock.Any())

// 	err := authService.SendCodeEmail(ctx, "email@ya.ru")
// 	require.NoError(t, err)
// }

// func TestUsersService_SignIn(t *testing.T) {
// 	authService, userRepo := mockAuthService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
// 		Return(
// 			repository.VerifyEmail{
// 				ExpiresAt: time.Now().Add(time.Minute),
// 			},
// 			nil,
// 		)
// 	userRepo.EXPECT().DeleteVerifyEmailById(ctx, gomock.Any())
// 	userRepo.EXPECT().GetUserByEmail(ctx, gomock.Any())
// 	userRepo.EXPECT().CreateSession(ctx, gomock.Any())

// 	res, err := authService.SignIn(ctx, service.UserSignInInput{})
// 	require.NoError(t, err)
// 	require.IsType(t, service.Tokens{}, res)
// }

func TestUsersService_SignInErrExpiredCode(t *testing.T) {
	authService, userRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any())

	res, err := authService.SignIn(ctx, domain.SignInRequest{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeExpired))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrCodeInvalid(t *testing.T) {
	authService, userRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(repository.VerifyEmail{}, repository.ErrRecordNotFound)

	res, err := authService.SignIn(ctx, domain.SignInRequest{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeInvalid))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrGetEmail(t *testing.T) {
	authService, userRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(repository.VerifyEmail{}, ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain.SignInRequest{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrDeleteEmail(t *testing.T) {
	authService, userRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(
			repository.VerifyEmail{
				ExpiresAt: time.Now().Add(time.Minute),
			},
			nil,
		)
	userRepo.EXPECT().DeleteVerifyEmailById(ctx, gomock.Any()).Return(ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain.SignInRequest{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrGetUser(t *testing.T) {
	authService, userRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(
			repository.VerifyEmail{
				ExpiresAt: time.Now().Add(time.Minute),
			},
			nil,
		)
	userRepo.EXPECT().DeleteVerifyEmailById(ctx, gomock.Any())
	userRepo.EXPECT().GetUserByEmail(ctx, gomock.Any()).
		Return(repository.User{}, ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain.SignInRequest{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
}

// func TestUsersService_SignInErrCreateSession(t *testing.T) {
// 	authService, userRepo := mockAuthService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
// 		Return(
// 			repository.VerifyEmail{
// 				ExpiresAt: time.Now().Add(time.Minute),
// 			},
// 			nil,
// 		)
// 	userRepo.EXPECT().DeleteVerifyEmailById(ctx, gomock.Any())
// 	userRepo.EXPECT().GetUserByEmail(ctx, gomock.Any())
// 	userRepo.EXPECT().CreateSession(ctx, gomock.Any()).
// 		Return(repository.Session{}, ErrInternalServerError)

// 	res, err := authService.SignIn(ctx, service.UserSignInInput{})
// 	require.True(t, errors.Is(err, ErrInternalServerError))
// 	require.IsType(t, service.Tokens{}, res)
// }

// func TestUsersService_RefreshToken(t *testing.T) {
// 	authService, userRepo := mockAuthService(t)

// 	duration := time.Minute
// 	userID, err := uuid.NewRandom()
// 	require.NoError(t, err)

// 	tokenManager, err := auth.NewPasetoManager(utils.RandomString(32))
// 	require.NoError(t, err)

// 	token, payload, err := tokenManager.CreateToken(userID, duration)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, payload)

// 	ctx := context.Background()
// 	userRepo.EXPECT().GetSession(ctx, gomock.Any())

// 	res, err := authService.RefreshToken(ctx, token)
// 	require.NoError(t, err)
// 	require.IsType(t, service.RefreshToken{}, res)
// }
