package service_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	domain_user "github.com/b0shka/backend/internal/domain/user"
	mock_repository "github.com/b0shka/backend/internal/repository/postgresql/mocks"
	"github.com/b0shka/backend/internal/service"
	mock_worker "github.com/b0shka/backend/internal/worker/mocks"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/identity"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var ErrInternalServerError = errors.New("test: internal server error")

func mockAuthService(t *testing.T) (
	*service.AuthService,
	*mock_repository.MockUsers,
	*mock_repository.MockSessions,
	*mock_repository.MockVerifyEmails,
) {
	repoCtl := gomock.NewController(t)
	defer repoCtl.Finish()

	workerCtl := gomock.NewController(t)
	defer workerCtl.Finish()

	repoUsers := mock_repository.NewMockUsers(repoCtl)
	repoSessions := mock_repository.NewMockSessions(repoCtl)
	repoVerifyEmails := mock_repository.NewMockVerifyEmails(repoCtl)
	worker := mock_worker.NewMockTaskDistributor(workerCtl)
	authService := service.NewAuthService(
		repoUsers,
		repoSessions,
		repoVerifyEmails,
		&hash.SHA256Hasher{},
		&auth.JWTManager{},
		&otp.TOTPGenerator{},
		&identity.IDGenerator{},
		config.AuthConfig{},
		worker,
	)

	return authService, repoUsers, repoSessions, repoVerifyEmails
}

// func TestUsersService_SendCodeEmail(t *testing.T) {
// 	authService, _, _, verifyEmailsRepo := mockAuthService(t)

// 	ctx := context.Background()
// 	verifyEmailsRepo.EXPECT().Create(ctx, gomock.Any())

// 	err := authService.SendCodeEmail(ctx, "email@ya.ru")
// 	require.NoError(t, err)
// }

// func TestUsersService_SignIn(t *testing.T) {
// 	authService, userRepo, sessionRepo, verifyEmailsRepo := mockAuthService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
// 		Return(
// 			domain_auth.VerifyEmail{
// 				ExpiresAt: time.Now().Add(time.Minute),
// 			},
// 			nil,
// 		)
// 	verifyEmailsRepo.EXPECT().DeleteById(ctx, gomock.Any())
// 	userRepo.EXPECT().GetByEmail(ctx, gomock.Any())
// 	sessionRepo.EXPECT().Create(ctx, gomock.Any())

// 	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
// 	require.NoError(t, err)
// 	require.IsType(t, domain_auth.SignInOutput{}, res)
// }

func TestUsersService_SignInErrExpiredCode(t *testing.T) {
	authService, _, _, verifyEmailsRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any())

	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
	require.True(t, errors.Is(err, domain.ErrSecretCodeExpired))
	require.IsType(t, domain_auth.SignInOutput{}, res)
}

// func TestUsersService_SignInErrCodeInvalid(t *testing.T) {
// 	authService, _, _, verifyEmailsRepo := mockAuthService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
// 		Return(domain_auth.VerifyEmail{}, repository.ErrRecordNotFound)

// 	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
// 	require.True(t, errors.Is(err, domain.ErrSecretCodeInvalid))
// 	require.IsType(t, domain_auth.SignInOutput{}, res)
// }

func TestUsersService_SignInErrGetEmail(t *testing.T) {
	authService, _, _, verifyEmailsRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
		Return(domain_auth.VerifyEmail{}, ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, domain_auth.SignInOutput{}, res)
}

func TestUsersService_SignInErrDeleteEmail(t *testing.T) {
	authService, _, _, verifyEmailsRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
		Return(
			domain_auth.VerifyEmail{
				ExpiresAt: time.Now().Add(time.Minute),
			},
			nil,
		)
	verifyEmailsRepo.EXPECT().DeleteByID(ctx, gomock.Any()).Return(ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, domain_auth.SignInOutput{}, res)
}

func TestUsersService_SignInErrGetUser(t *testing.T) {
	authService, userRepo, _, verifyEmailsRepo := mockAuthService(t)

	// ctx := context.Background()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
		Return(
			domain_auth.VerifyEmail{
				ExpiresAt: time.Now().Add(time.Minute),
			},
			nil,
		)
	verifyEmailsRepo.EXPECT().DeleteByID(ctx, gomock.Any())
	userRepo.EXPECT().GetByEmail(ctx, gomock.Any()).
		Return(domain_user.User{}, ErrInternalServerError)

	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, domain_auth.SignInOutput{}, res)
}

// func TestUsersService_SignInErrCreateSession(t *testing.T) {
// 	authService, userRepo, sessionRepo, verifyEmailsRepo := mockAuthService(t)

// 	// ctx := context.Background()
// 	w := httptest.NewRecorder()
// 	ctx, _ := gin.CreateTestContext(w)

// 	verifyEmailsRepo.EXPECT().Get(ctx, gomock.Any()).
// 		Return(
// 			domain_auth.VerifyEmail{
// 				ExpiresAt: time.Now().Add(time.Minute),
// 			},
// 			nil,
// 		)
// 	verifyEmailsRepo.EXPECT().DeleteByID(ctx, gomock.Any())
// 	userRepo.EXPECT().GetByEmail(ctx, gomock.Any())
// 	sessionRepo.EXPECT().Create(ctx, gomock.Any()).
// 		Return(domain_auth.Session{}, ErrInternalServerError)

// 	res, err := authService.SignIn(ctx, domain_auth.SignInInput{})
// 	require.True(t, errors.Is(err, ErrInternalServerError))
// 	require.IsType(t, domain_auth.SignInOutput{}, res)
// }

func TestUsersService_RefreshToken(t *testing.T) {
	authService, _, sessionRepo, _ := mockAuthService(t)

	duration := time.Minute
	userID, err := uuid.NewRandom()
	require.NoError(t, err)

	symmetricKey, err := utils.RandomString(32)
	require.NoError(t, err)
	tokenManager, err := auth.NewPasetoManager(symmetricKey)
	require.NoError(t, err)

	token, payload, err := tokenManager.CreateToken(userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	ctx := context.Background()
	sessionRepo.EXPECT().Get(ctx, gomock.Any())

	res, _ := authService.RefreshToken(ctx, domain_auth.RefreshTokenInput{
		RefreshToken: token,
	})
	// require.NoError(t, err)
	require.IsType(t, domain_auth.RefreshTokenOutput{}, res)
}
