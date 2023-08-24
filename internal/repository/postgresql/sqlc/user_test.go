package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/b0shka/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomUser() {

}

func TestRepository_CreateUser(t *testing.T) {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := CreateUserParams{
		ID:       id,
		Email:    fmt.Sprintf("%s@ya.ru", utils.RandomString(7)),
		Username: utils.RandomString(10),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Photo, user.Photo)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
}

func TestRepository_GetUserById(t *testing.T) {

}
