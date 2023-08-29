package service_test

import (
	"context"
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
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var ErrInternalServerError = errors.New("test: internal server error")

func mockUserService(t *testing.T) (*service.UsersService, *mock_repository.MockStore) {
	repoCtl := gomock.NewController(t)
	defer repoCtl.Finish()

	workerCtl := gomock.NewController(t)
	defer workerCtl.Finish()

	repo := mock_repository.NewMockStore(repoCtl)
	worker := mock_worker.NewMockTaskDistributor(workerCtl)
	userService := service.NewUsersService(
		repo,
		&hash.SHA256Hasher{},
		&auth.JWTManager{},
		config.AuthConfig{},
		worker,
	)

	return userService, repo
}

// func TestUsersService_SendCodeEmail(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

// 	ctx := context.Background()
// 	userRepo.EXPECT().CreateVerifyEmail(ctx, gomock.Any())

// 	err := userService.SendCodeEmail(ctx, "email@ya.ru")
// 	require.NoError(t, err)
// }

// func TestUsersService_SignIn(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

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

// 	res, err := userService.SignIn(ctx, service.UserSignInInput{})
// 	require.NoError(t, err)
// 	require.IsType(t, service.Tokens{}, res)
// }

func TestUsersService_SignInErrExpiredCode(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any())

	user, res, err := userService.SignIn(ctx, domain.UserSignIn{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeExpired))
	require.IsType(t, service.Tokens{}, res)
	require.IsType(t, repository.User{}, user)
}

func TestUsersService_SignInErrCodeInvalid(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(repository.VerifyEmail{}, repository.ErrRecordNotFound)

	user, res, err := userService.SignIn(ctx, domain.UserSignIn{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeInvalid))
	require.IsType(t, service.Tokens{}, res)
	require.IsType(t, repository.User{}, user)
}

func TestUsersService_SignInErrGetEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any()).
		Return(repository.VerifyEmail{}, ErrInternalServerError)

	user, res, err := userService.SignIn(ctx, domain.UserSignIn{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
	require.IsType(t, repository.User{}, user)
}

func TestUsersService_SignInErrDeleteEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

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

	user, res, err := userService.SignIn(ctx, domain.UserSignIn{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
	require.IsType(t, repository.User{}, user)
}

func TestUsersService_SignInErrGetUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

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

	user, res, err := userService.SignIn(ctx, domain.UserSignIn{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, service.Tokens{}, res)
	require.IsType(t, repository.User{}, user)
}

// func TestUsersService_SignInErrCreateSession(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

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

// 	res, err := userService.SignIn(ctx, service.UserSignInInput{})
// 	require.True(t, errors.Is(err, ErrInternalServerError))
// 	require.IsType(t, service.Tokens{}, res)
// }

// func TestUsersService_RefreshToken(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

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

// 	res, err := userService.RefreshToken(ctx, token)
// 	require.NoError(t, err)
// 	require.IsType(t, service.RefreshToken{}, res)
// }

func TestUsersService_Get(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetUserById(ctx, gomock.Any())

	res, err := userService.GetByID(ctx, uuid.UUID{})
	require.NoError(t, err)
	require.IsType(t, repository.User{}, res)
}

func TestUsersService_GetErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetUserById(ctx, gomock.Any()).Return(repository.User{}, ErrInternalServerError)

	res, err := userService.GetByID(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, repository.User{}, res)
}

func TestUsersService_Update(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().UpdateUser(ctx, gomock.Any())

	err := userService.Update(ctx, uuid.UUID{}, domain.UserUpdate{})
	require.NoError(t, err)
}

func TestUsersService_UpdateErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().UpdateUser(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Update(ctx, uuid.UUID{}, domain.UserUpdate{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_Delete(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().DeleteSession(ctx, gomock.Any())
	userRepo.EXPECT().GetUserById(ctx, gomock.Any())
	userRepo.EXPECT().DeleteVerifyEmailByEmail(ctx, gomock.Any())
	userRepo.EXPECT().DeleteUser(ctx, gomock.Any())

	err := userService.Delete(ctx, uuid.UUID{})
	require.NoError(t, err)
}

func TestUsersService_DeleteErrDelSession(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().DeleteSession(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrGetUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().DeleteSession(ctx, gomock.Any())
	userRepo.EXPECT().GetUserById(ctx, gomock.Any()).
		Return(repository.User{}, ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrDelVerEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().DeleteSession(ctx, gomock.Any())
	userRepo.EXPECT().GetUserById(ctx, gomock.Any())
	userRepo.EXPECT().DeleteVerifyEmailByEmail(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrDelUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().DeleteSession(ctx, gomock.Any())
	userRepo.EXPECT().GetUserById(ctx, gomock.Any())
	userRepo.EXPECT().DeleteVerifyEmailByEmail(ctx, gomock.Any())
	userRepo.EXPECT().DeleteUser(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}
