package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	formatTimeLayout = "Jan _2, 2006 15:04:05 (MST)"
)

type AuthService struct {
	repo            repository.Store
	hasher          hash.Hasher
	tokenManager    auth.Manager
	otpGenerator    otp.Generator
	authConfig      config.AuthConfig
	taskDistributor worker.TaskDistributor
}

func NewAuthService(
	repo repository.Store,
	hasher hash.Hasher,
	tokenManager auth.Manager,
	otpGenerator otp.Generator,
	authConfig config.AuthConfig,
	taskDistributor worker.TaskDistributor,
) *AuthService {
	return &AuthService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		otpGenerator:    otpGenerator,
		authConfig:      authConfig,
		taskDistributor: taskDistributor,
	}
}

func (s *AuthService) SendCodeEmail(ctx context.Context, email string) error {
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
		return err
	} else if errors.Is(err, repository.ErrRecordNotFound) {
		userID, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		_, err = s.repo.CreateUser(ctx, repository.CreateUserParams{
			ID:    userID,
			Email: email,
		})
		if err != nil {
			return err
		}
	}

	secretCode := s.otpGenerator.RandomCode(s.authConfig.VerificationCodeLength)
	taskPayload := &worker.PayloadSendVerifyEmail{
		Email:      email,
		SecretCode: secretCode,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = s.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)

	return err
}

func (s *AuthService) SignIn(ctx *gin.Context, inp domain.SignInRequest) (Tokens, error) {
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
		if errors.Is(err, repository.ErrRecordNotFound) {
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

	tokens, err := s.createSession(ctx, user.ID)
	if err != nil {
		return Tokens{}, err
	}

	taskPayload := &worker.PayloadSendLoginNotification{
		Email:     inp.Email,
		UserAgent: ctx.Request.UserAgent(),
		ClientIP:  ctx.ClientIP(),
		Time:      time.Now().Format(formatTimeLayout),
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(5 * time.Second),
		asynq.Queue(worker.QueueDefault),
	}

	err = s.taskDistributor.DistributeTaskSendLoginNotification(ctx, taskPayload, opts...)
	if err != nil {
		return Tokens{}, err
	}

	return tokens, nil
}

func (s *AuthService) createSession(ctx *gin.Context, id uuid.UUID) (Tokens, error) {
	var res Tokens

	refreshToken, refreshPayload, err := s.tokenManager.CreateToken(
		id,
		s.authConfig.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return res, err
	}

	res.SessionID = refreshPayload.ID
	res.RefreshToken = refreshToken

	accessToken, _, err := s.tokenManager.CreateToken(
		id,
		s.authConfig.JWT.AccessTokenTTL,
	)
	if err != nil {
		return res, err
	}

	res.AccessToken = accessToken

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

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (RefreshToken, error) {
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

	accessToken, _, err := s.tokenManager.CreateToken(
		refreshPayload.UserID,
		s.authConfig.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return RefreshToken{}, err
	}

	res.AccessToken = accessToken

	return res, nil
}
