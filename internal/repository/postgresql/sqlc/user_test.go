package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/b0shka/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	arg := CreateUserParams{
		ID:       id,
		Email:    fmt.Sprintf("%s@ya.ru", utils.RandomString(7)),
		Username: utils.RandomString(10),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Photo, user.Photo)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestRepository_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestRepository_GetUserById(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUserById(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Photo, user2.Photo)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestRepository_GetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testStore.GetUserByEmail(context.Background(), user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Photo, user2.Photo)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestRepository_UpdateUser(t *testing.T) {
	user := createRandomUser(t)

	arg := UpdateUserParams{
		ID:       user.ID,
		Username: utils.RandomString(10),
		Photo: pgtype.Text{
			String: fmt.Sprintf("https://%s.png", utils.RandomString(7)),
			Valid:  true,
		},
	}

	err := testStore.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
}

func TestRepository_DeleteUser(t *testing.T) {
	user := createRandomUser(t)
	err := testStore.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)
}
