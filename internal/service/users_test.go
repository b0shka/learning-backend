package service_test

import (
	"context"
	"errors"
	"testing"

	domain_user "github.com/b0shka/backend/internal/domain/user"
	mock_repository "github.com/b0shka/backend/internal/repository/postgresql/mocks"
	"github.com/b0shka/backend/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func mockUserService(t *testing.T) (
	*service.UsersService,
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
	userService := service.NewUsersService(
		repoUsers,
		repoSessions,
		repoVerifyEmails,
	)

	return userService, repoUsers, repoSessions, repoVerifyEmails
}

func TestUsersService_Get(t *testing.T) {
	userService, userRepo, _, _ := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetByID(ctx, gomock.Any())

	res, err := userService.GetByID(ctx, uuid.UUID{})
	require.NoError(t, err)
	require.IsType(t, domain_user.User{}, res)
}

func TestUsersService_GetErr(t *testing.T) {
	userService, userRepo, _, _ := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().GetByID(ctx, gomock.Any()).Return(domain_user.User{}, ErrInternalServerError)

	res, err := userService.GetByID(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
	require.IsType(t, domain_user.User{}, res)
}

func TestUsersService_Delete(t *testing.T) {
	userService, userRepo, sessionRepo, verifyEmailsRepo := mockUserService(t)

	ctx := context.Background()
	sessionRepo.EXPECT().Delete(ctx, gomock.Any())
	userRepo.EXPECT().GetByID(ctx, gomock.Any())
	verifyEmailsRepo.EXPECT().DeleteByEmail(ctx, gomock.Any())
	userRepo.EXPECT().Delete(ctx, gomock.Any())

	err := userService.Delete(ctx, uuid.UUID{})
	require.NoError(t, err)
}

func TestUsersService_DeleteErrDelSession(t *testing.T) {
	userService, _, sessionRepo, _ := mockUserService(t)

	ctx := context.Background()
	sessionRepo.EXPECT().Delete(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrGetUser(t *testing.T) {
	userService, userRepo, sessionRepo, _ := mockUserService(t)

	ctx := context.Background()
	sessionRepo.EXPECT().Delete(ctx, gomock.Any())
	userRepo.EXPECT().GetByID(ctx, gomock.Any()).
		Return(domain_user.User{}, ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrDelVerEmail(t *testing.T) {
	userService, userRepo, sessionRepo, verifyEmailsRepo := mockUserService(t)

	ctx := context.Background()
	sessionRepo.EXPECT().Delete(ctx, gomock.Any())
	userRepo.EXPECT().GetByID(ctx, gomock.Any())
	verifyEmailsRepo.EXPECT().DeleteByEmail(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}

func TestUsersService_DeleteErrDelUser(t *testing.T) {
	userService, userRepo, sessionRepo, verifyEmailsRepo := mockUserService(t)

	ctx := context.Background()
	sessionRepo.EXPECT().Delete(ctx, gomock.Any())
	userRepo.EXPECT().GetByID(ctx, gomock.Any())
	verifyEmailsRepo.EXPECT().DeleteByEmail(ctx, gomock.Any())
	userRepo.EXPECT().Delete(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Delete(ctx, uuid.UUID{})
	require.True(t, errors.Is(err, ErrInternalServerError))
}
