package repository

import (
	"context"
	"testing"
	"time"

	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	domain_user "github.com/b0shka/backend/internal/domain/user"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomVerifyEmail(t *testing.T, user domain_user.User) domain_auth.VerifyEmail {
	verifyEmailID, err := uuid.NewRandom()
	require.NoError(t, err)

	salt, err := utils.RandomString(32)
	require.NoError(t, err)

	hasher, err := hash.NewSHA256Hasher(salt)
	require.NoError(t, err)

	otpGenerator := otp.NewTOTPGenerator()

	code := otpGenerator.RandomCode(6)
	codeHash, err := hasher.HashCode(code)
	require.NoError(t, err)

	arg := CreateVerifyEmailParams{
		ID:         verifyEmailID,
		Email:      user.Email,
		SecretCode: codeHash,
		ExpiresAt:  time.Now().Add(time.Minute * 5),
	}

	verifyEmail, err := testRepos.VerifyEmails.Create(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail)

	return verifyEmail
}

func TestRepository_CreateVerifyEmail(t *testing.T) {
	user := createRandomUser(t)
	createRandomVerifyEmail(t, user)
}

func TestRepository_GetVerifyEmail(t *testing.T) {
	user := createRandomUser(t)
	verifyEmail1 := createRandomVerifyEmail(t, user)

	arg := GetVerifyEmailParams{
		Email:      verifyEmail1.Email,
		SecretCode: verifyEmail1.SecretCode,
	}

	verifyEmail2, err := testRepos.VerifyEmails.Get(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail2)

	require.Equal(t, verifyEmail1.ID, verifyEmail2.ID)
	require.Equal(t, verifyEmail1.Email, verifyEmail2.Email)
	require.Equal(t, verifyEmail1.SecretCode, verifyEmail2.SecretCode)
	require.WithinDuration(t, verifyEmail1.ExpiresAt, verifyEmail2.ExpiresAt, time.Second)
}

func TestRepository_DeleteVerifyEmailById(t *testing.T) {
	user := createRandomUser(t)
	verifyEmail := createRandomVerifyEmail(t, user)
	err := testRepos.VerifyEmails.DeleteByID(context.Background(), verifyEmail.ID)
	require.NoError(t, err)
}

func TestRepository_DeleteVerifyEmailByEmail(t *testing.T) {
	user := createRandomUser(t)
	createRandomVerifyEmail(t, user)
	err := testRepos.VerifyEmails.DeleteByEmail(context.Background(), user.Email)
	require.NoError(t, err)
}
