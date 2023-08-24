package service_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	mock_repository "github.com/b0shka/backend/internal/repository/mocks"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/gin-gonic/gin"
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
		&auth.JWTManager{},
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

// func TestUsersService_SignIn(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
// 		Return(
// 			domain.VerifyEmail{
// 				ExpiresAt: time.Now().Unix(),
// 			},
// 			nil,
// 		)
// 	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
// 	userRepo.EXPECT().GetUser(ctx, gomock.Any())
// 	userRepo.EXPECT().CreateUser(ctx, gomock.Any())
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

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any())

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeExpired))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrGetEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(domain.VerifyEmail{}, errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})

	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrRemoveEmail(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiresAt: time.Now().Unix(),
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

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiresAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
	userRepo.EXPECT().GetUser(ctx, gomock.Any()).Return(domain.User{}, errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

func TestUsersService_SignInErrCreateUser(t *testing.T) {
	userService, userRepo := mockUserService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	userRepo.EXPECT().GetVerifyEmail(ctx, gomock.Any(), gomock.Any()).
		Return(
			domain.VerifyEmail{
				ExpiresAt: time.Now().Unix(),
			},
			nil,
		)
	userRepo.EXPECT().RemoveVerifyEmail(ctx, gomock.Any())
	userRepo.EXPECT().GetUser(ctx, gomock.Any()).Return(domain.User{}, domain.ErrUserNotFound)
	userRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(errInternalServErr)

	res, err := userService.SignIn(ctx, service.UserSignInInput{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, service.Tokens{}, res)
}

// func TestUsersService_RefreshToken(t *testing.T) {
// 	userService, userRepo := mockUserService(t)

// 	userId := primitive.NewObjectID()
// 	duration := time.Minute

// 	tokenManager, err := auth.NewPasetoManager(utils.RandomString(32))
// 	require.NoError(t, err)

// 	token, payload, err := tokenManager.CreateToken(userId, duration)
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
	userRepo.EXPECT().GetUser(ctx, gomock.Any())

	res, err := userService.Get(ctx, primitive.ObjectID{})
	require.NoError(t, err)
	require.IsType(t, domain.User{}, res)
}

func TestUsersService_GetErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetUser(ctx, gomock.Any()).Return(domain.User{}, errInternalServErr)

	res, err := userService.Get(ctx, primitive.ObjectID{})
	require.True(t, errors.Is(err, errInternalServErr))
	require.IsType(t, domain.User{}, res)
}

func TestUsersService_Update(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().UpdateUser(ctx, gomock.Any(), gomock.Any())

	err := userService.Update(ctx, primitive.NewObjectID(), domain.UserUpdate{})
	require.NoError(t, err)
}

func TestUsersService_UpdateErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().UpdateUser(ctx, gomock.Any(), gomock.Any()).Return(errInternalServErr)

	err := userService.Update(ctx, primitive.NewObjectID(), domain.UserUpdate{})
	require.True(t, errors.Is(err, errInternalServErr))
}
