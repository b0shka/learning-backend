package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomSession(t *testing.T, user User) Session {
	sessionID, err := uuid.NewRandom()
	require.NoError(t, err)

	tokenManager, err := auth.NewPasetoManager(utils.RandomString(32))
	require.NoError(t, err)

	refreshToken, _, err := tokenManager.CreateToken(user.ID, time.Hour)
	require.NoError(t, err)

	arg := CreateSessionParams{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    utils.RandomString(20),
		ClientIp:     fmt.Sprintf("%d:%d:%d:%d", utils.RandomInt(1, 255), utils.RandomInt(1, 255), utils.RandomInt(1, 255), utils.RandomInt(1, 255)),
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	session, err := testStore.CreateSession(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, session)

	return session
}

func TestRepository_CreateSession(t *testing.T) {
	user := createRandomUser(t)
	createRandomSession(t, user)
}

func TestRepository_GetSession(t *testing.T) {
	user := createRandomUser(t)
	session1 := createRandomSession(t, user)
	session2, err := testStore.GetSession(context.Background(), session1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, session2)

	require.Equal(t, session1.ID, session2.ID)
	require.Equal(t, session1.UserID, session2.UserID)
	require.Equal(t, session1.RefreshToken, session2.RefreshToken)
	require.Equal(t, session1.UserAgent, session2.UserAgent)
	require.Equal(t, session1.ClientIp, session2.ClientIp)
	require.Equal(t, session1.IsBlocked, session2.IsBlocked)
	require.WithinDuration(t, session1.ExpiresAt, session2.ExpiresAt, time.Second)
}

func TestRepository_DeleteSession(t *testing.T) {
	user := createRandomUser(t)
	createRandomSession(t, user)
	err := testStore.DeleteSession(context.Background(), user.ID)
	require.NoError(t, err)
}
