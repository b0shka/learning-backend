package service

import (
	"context"

	domain_user "github.com/b0shka/backend/internal/domain/user"
	repository "github.com/b0shka/backend/internal/repository/postgresql"
	"github.com/google/uuid"
)

type UsersService struct {
	repoUsers        repository.Users
	repoSessions     repository.Sessions
	repoVerifyEmails repository.VerifyEmails
}

func NewUsersService(
	repoUsers repository.Users,
	repoSessions repository.Sessions,
	repoVerifyEmails repository.VerifyEmails,
) *UsersService {
	return &UsersService{
		repoUsers:        repoUsers,
		repoSessions:     repoSessions,
		repoVerifyEmails: repoVerifyEmails,
	}
}

func (s *UsersService) GetByID(ctx context.Context, id uuid.UUID) (domain_user.User, error) {
	return s.repoUsers.GetByID(ctx, id)
}

func (s *UsersService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repoSessions.Delete(ctx, id)
	if err != nil {
		return err
	}

	user, err := s.repoUsers.GetByID(ctx, id)
	if err != nil {
		return err
	}

	err = s.repoVerifyEmails.DeleteByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	return s.repoUsers.Delete(ctx, id)
}
