package repository

import (
	"context"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/domain"
	mock_repository "github.com/b0shka/backend/internal/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func mockRepository(t *testing.T) *mock_repository.MockUsers {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	userRepo := mock_repository.NewMockUsers(mockCtl)
	return userRepo
}

func TestAddVerifyEmail(t *testing.T) {
	userRepo := mockRepository(t)
	ctx := context.Background()

	timestamp := time.Now().Unix()
	in := []domain.VerifyEmail{
		{
			Email:      "email@ya.ru",
			SecretCode: "543234",
			ExpiredAt:  timestamp + 300,
		},
	}

	for _, verifyEmail := range in {
		userRepo.EXPECT().AddVerifyEmail(ctx, verifyEmail).Return(nil).Times(1)
		err := userRepo.AddVerifyEmail(ctx, verifyEmail)
		require.NoError(t, err)
	}
}

type UserSignInInput struct {
	Email      string
	SecretCode int32
}

func TestGetVerifyEmail(t *testing.T) {
	userRepo := mockRepository(t)
	ctx := context.Background()

	in := []UserSignInInput{
		{
			Email:      "email@ya.ru",
			SecretCode: 543234,
		},
	}

	timestamp := time.Now().Unix()
	expResp := []domain.VerifyEmail{
		{
			Email:      "email@ya.ru",
			SecretCode: "543234",
			ExpiredAt:  timestamp + 300,
		},
	}

	for index, authCode := range in {
		userRepo.EXPECT().GetVerifyEmail(ctx, authCode.Email, string(authCode.SecretCode)).Return(expResp[index], nil).Times(1)
		verifyEmail, err := userRepo.GetVerifyEmail(ctx, authCode.Email, string(authCode.SecretCode))

		require.NoError(t, err)
		require.Equal(t, expResp[index], verifyEmail)
	}
}

func TestCreateUser(t *testing.T) {
	usersRepo := mockRepository(t)
	ctx := context.Background()

	timestamp := time.Now().Unix()
	in := []domain.User{
		{
			Email:     "email@ya.ru",
			Photo:     "",
			Name:      "Vanya",
			CreatedAt: timestamp,
		},
	}

	for _, user := range in {
		usersRepo.EXPECT().Create(ctx, user).Return(nil).Times(1)
		err := usersRepo.Create(ctx, user)
		require.NoError(t, err)
	}
}

func TestGetUser(t *testing.T) {
	usersRepo := mockRepository(t)
	ctx := context.Background()

	in := []interface{}{
		"email@ya.ru",
		primitive.NewObjectID(),
	}

	timestamp := time.Now().Unix()
	expResp := domain.User{
		ID:        primitive.ObjectID{},
		Email:     "email@ya.ru",
		Photo:     "",
		Name:      "Vanya",
		CreatedAt: timestamp,
	}

	for _, email := range in {
		usersRepo.EXPECT().Get(ctx, email).Return(expResp, nil).Times(1)
		user, err := usersRepo.Get(ctx, email)

		require.NoError(t, err)
		require.Equal(t, expResp, user)
	}
}
