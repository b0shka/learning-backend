package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	mock_repository "github.com/b0shka/backend/internal/repository/mocks"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/b0shka/backend/pkg/auth"
	"github.com/golang/mock/gomock"
)

var errInternalServErr = errors.New("test: internal server error")

func mockUserService(t *testing.T) (*service.UsersService, *mock_repository.MockUsers) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	userRepo := mock_repository.NewMockUsers(mockCtl)
	userService := service.NewUsersService(
		userRepo,
		&hash.SHA256Hasher{},
		&auth.Manager{},
		email.EmailService{},
		config.EmailConfig{},
		config.AuthConfig{},
	)

	return userService, userRepo
}

// func TestUsersService_SendCodeEmail(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

// 	ctx := context.Background()
// 	userRepo.EXPECT().AddVerifyEmail(ctx, gomock.Any())

// 	err := userService.SendCodeEmail(ctx, "email@ya.ru")
// 	assert.NoError(t, err)
// }

func TestUsersService_SignIn(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiredAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
	userRepo.EXPECT().Get(ctx, gomock.Any())
	userRepo.EXPECT().Create(ctx, gomock.Any())

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.NoError(t, err)
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrExpiredCode(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any())

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeExpired))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrGetEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(domain.VerifyEmail{}, errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})

	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrRemoveEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiredAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any()).Return(errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrGetUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiredAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
	userRepo.EXPECT().Get(ctx, gomock.Any()).Return(domain.User{}, errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrCreateUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiredAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
	userRepo.EXPECT().Get(ctx, gomock.Any()).Return(domain.User{}, domain.ErrUserNotFound)
	userRepo.EXPECT().Create(ctx, gomock.Any()).Return(errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_Get(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().Get(ctx, gomock.Any())

	res, err := userService.Get(ctx, primitive.ObjectID{})
	require.NoError(t, err)
	require.IsType(t, domain.User{}, res)
}

func TestUsersService_GetErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().Get(ctx, gomock.Any()).Return(domain.User{}, errInternalServErr)

	res, err := userService.Get(ctx, primitive.ObjectID{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, domain.User{}, res)
}

func TestUsersService_Update(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().Update(ctx, gomock.Any())

	err := userService.Update(ctx, domain.UserUpdate{})
	require.NoError(t, err)
}

func TestUsersService_UpdateErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().Update(ctx, gomock.Any()).Return(errInternalServErr)

	err := userService.Update(ctx, domain.UserUpdate{})
	require.True(t, errors.Is(err, errInternalServErr))
}
