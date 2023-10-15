package service

import (
	"context"

	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UsersService struct {
	repo repository.Store
}

func NewUsersService(
	repo repository.Store,
) *UsersService {
	return &UsersService{
		repo: repo,
	}
}

func (s *UsersService) GetByID(ctx context.Context, id uuid.UUID) (repository.User, error) {
	return s.repo.GetUserById(ctx, id)
}

func (s *UsersService) Update(ctx context.Context, id uuid.UUID, user domain.UpdateUserRequest) error {
	arg := repository.UpdateUserParams{
		ID:       id,
		Username: user.Username,
		Photo: pgtype.Text{
			String: user.Photo,
			Valid:  true,
		},
	}

	return s.repo.UpdateUser(ctx, arg)
}

func (s *UsersService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteSession(ctx, id)
	if err != nil {
		return err
	}

	user, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	err = s.repo.DeleteVerifyEmailByEmail(ctx, user.Email)
	if err != nil {
		return err
	}

	return s.repo.DeleteUser(ctx, id)
}
