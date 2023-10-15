package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/b0shka/backend/internal/domain"
	mock_repository "github.com/b0shka/backend/internal/repository/postgresql/mocks"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/service"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func mockUserService(t *testing.T) (*service.UsersService, *mock_repository.MockStore) {
	repoCtl := gomock.NewController(t)
	defer repoCtl.Finish()

	workerCtl := gomock.NewController(t)
	defer workerCtl.Finish()

	repo := mock_repository.NewMockStore(repoCtl)
	userService := service.NewUsersService(
		repo,
	)

	return userService, repo
}

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

	err := userService.Update(ctx, uuid.UUID{}, domain.UpdateUserRequest{})
	require.NoError(t, err)
}

func TestUsersService_UpdateErr(t *testing.T) {
	userService, userRepo := mockUserService(t)

	ctx := context.Background()
	userRepo.EXPECT().UpdateUser(ctx, gomock.Any()).Return(ErrInternalServerError)

	err := userService.Update(ctx, uuid.UUID{}, domain.UpdateUserRequest{})
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
