package service

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UsersService struct {
	repo         *repository.Store
	hasher       hash.Hasher
	tokenManager auth.Manager
	emailService email.EmailService
	emailConfig  config.EmailConfig
	authConfig   config.AuthConfig
}

func NewUsersService(
	repo *repository.Store,
	hasher hash.Hasher,
	tokenManager auth.Manager,
	emailService email.EmailService,
	emailConfig config.EmailConfig,
	authConfig config.AuthConfig,
) *UsersService {
	return &UsersService{
		repo:         repo,
		hasher:       hasher,
		tokenManager: tokenManager,
		emailService: emailService,
		emailConfig:  emailConfig,
		authConfig:   authConfig,
	}
}

func (s *UsersService) SendCodeEmail(ctx context.Context, email string) error {
	secretCode := utils.RandomInt(100000, 999999)
	secretCodeStr := strconv.Itoa(int(secretCode))
	secretCodeHash, err := s.hasher.HashCode(secretCodeStr)
	if err != nil {
		return err
	}

	var content bytes.Buffer
	contentHtml, err := template.ParseFiles(s.emailConfig.Templates.Verify)
	if err != nil {
		return err
	}

	err = contentHtml.Execute(&content, UserSignInInput{
		Email:      email,
		SecretCode: secretCode,
	})
	if err != nil {
		return err
	}

	err = s.emailService.SendEmail(domain.VerifyEmailConfig{
		Subject: s.emailConfig.Subjects.Verify,
		Content: content.String(),
	}, email)
	if err != nil {
		return err
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userID, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			user, err = s.repo.CreateUser(ctx, repository.CreateUserParams{
				ID:    userID,
				Email: email,
			})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	verifyEmailID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	verifyEmail := repository.CreateVerifyEmailParams{
		ID:         verifyEmailID,
		Email:      user.Email,
		SecretCode: secretCodeHash,
		ExpiresAt:  time.Now().Add(s.authConfig.SercetCodeLifetime),
	}

	_, err = s.repo.CreateVerifyEmail(ctx, verifyEmail)
	return err
}

func (s *UsersService) SignIn(ctx *gin.Context, inp UserSignInInput) (Tokens, error) {
	secretCodeStr := strconv.Itoa(int(inp.SecretCode))
	secretCodeHash, err := s.hasher.HashCode(secretCodeStr)
	if err != nil {
		return Tokens{}, err
	}

	arg := repository.GetVerifyEmailParams{
		Email:      inp.Email,
		SecretCode: secretCodeHash,
	}

	verifyEmail, err := s.repo.GetVerifyEmail(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tokens{}, domain.ErrSecretCodeInvalid
		}
		return Tokens{}, err
	}

	if time.Now().After(verifyEmail.ExpiresAt) {
		return Tokens{}, domain.ErrSecretCodeExpired
	}

	err = s.repo.DeleteVerifyEmailById(ctx, verifyEmail.ID)
	if err != nil {
		return Tokens{}, err
	}

	user, err := s.repo.GetUserByEmail(ctx, inp.Email)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *UsersService) createSession(ctx *gin.Context, id uuid.UUID) (Tokens, error) {
	var res Tokens

	refreshToken, refreshPayload, err := s.tokenManager.CreateToken(
		id,
		s.authConfig.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return res, err
	}

	res.RefreshToken = refreshToken
	res.RefreshTokenExpiresAt = refreshPayload.ExpiresAt

	accessToken, accessPayload, err := s.tokenManager.CreateToken(
		id,
		s.authConfig.JWT.AccessTokenTTL,
	)
	if err != nil {
		return res, err
	}

	res.AccessToken = accessToken
	res.AccessTokenExpiresAt = accessPayload.ExpiresAt

	session := repository.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       id,
		RefreshToken: res.RefreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	}

	if _, err := s.repo.CreateSession(ctx, session); err != nil {
		return Tokens{}, err
	}

	return res, nil
}

func (s *UsersService) RefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error) {
	var res RefreshToken

	refreshPayload, err := s.tokenManager.VerifyToken(refreshToken)
	if err != nil {
		return RefreshToken{}, err
	}

	session, err := s.repo.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		return RefreshToken{}, err
	}

	if session.IsBlocked {
		return RefreshToken{}, domain.ErrSessionBlocked
	}

	if refreshPayload.UserID != session.UserID {
		return RefreshToken{}, domain.ErrIncorrectSessionUser
	}

	if refreshToken != session.RefreshToken {
		return RefreshToken{}, domain.ErrMismatchedSession
	}

	if time.Now().After(session.ExpiresAt) {
		return RefreshToken{}, domain.ErrExpiredToken
	}

	accessToken, accessPayload, err := s.tokenManager.CreateToken(
		refreshPayload.UserID,
		s.authConfig.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return RefreshToken{}, err
	}

	res.AccessToken = accessToken
	res.AccessTokenExpiresAt = accessPayload.ExpiresAt

	return res, nil
}

func (s *UsersService) GetById(ctx context.Context, id uuid.UUID) (repository.User, error) {
	return s.repo.GetUserById(ctx, id)
}

func (s *UsersService) Update(ctx context.Context, id uuid.UUID, user domain.UserUpdate) error {
	arg := repository.UpdateUserParams{
		ID:       id,
		Username: user.Username,
		Photo:    user.Photo,
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
