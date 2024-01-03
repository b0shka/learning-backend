package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain_user "github.com/b0shka/backend/internal/domain/user"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) domain_user.User {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	email, err := utils.RandomString(7)
	require.NoError(t, err)

	arg := CreateUserParams{
		ID:    id,
		Email: fmt.Sprintf("%s@ya.ru", email),
	}

	user, err := testRepos.Users.Create(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestRepository_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestRepository_GetUserById(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testRepos.Users.GetByID(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestRepository_GetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testRepos.Users.GetByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestRepository_DeleteUser(t *testing.T) {
	user := createRandomUser(t)
	err := testRepos.Users.Delete(context.Background(), user.ID)
	require.NoError(t, err)
}
