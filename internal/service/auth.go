package service

import (
	"context"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	repository "github.com/b0shka/backend/internal/repository/postgresql"
	"github.com/b0shka/backend/internal/worker"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/identity"
	"github.com/b0shka/backend/pkg/otp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	formatTimeLayout = "Jan _2, 2006 15:04:05 (MST)"
)

type AuthService struct {
	repoUsers        repository.Users
	repoSessions     repository.Sessions
	repoVerifyEmails repository.VerifyEmails
	hasher           hash.Hasher
	tokenManager     auth.Manager
	otpGenerator     otp.Generator
	idGenerator      identity.Generator
	authConfig       config.AuthConfig
	taskDistributor  worker.TaskDistributor
}

func NewAuthService(
	repoUsers repository.Users,
	repoSessions repository.Sessions,
	repoVerifyEmails repository.VerifyEmails,
	hasher hash.Hasher,
	tokenManager auth.Manager,
	otpGenerator otp.Generator,
	idGenerator identity.Generator,
	authConfig config.AuthConfig,
	taskDistributor worker.TaskDistributor,
) *AuthService {
	return &AuthService{
		repoVerifyEmails: repoVerifyEmails,
		repoSessions:     repoSessions,
		repoUsers:        repoUsers,
		hasher:           hasher,
		tokenManager:     tokenManager,
		otpGenerator:     otpGenerator,
		idGenerator:      idGenerator,
		authConfig:       authConfig,
		taskDistributor:  taskDistributor,
	}
}

func (s *AuthService) SendCodeEmail(ctx context.Context, inp domain_auth.SendCodeEmailInput) error {
	userParams := repository.CreateUserParams{
		ID:    s.idGenerator.GenerateUUID(),
		Email: inp.Email,
	}

	_, err := s.repoUsers.Create(ctx, userParams)
	if err != nil {
		return err
	}

	secretCode := s.otpGenerator.RandomCode(s.authConfig.VerificationCodeLength)
	taskPayload := &worker.PayloadSendVerifyEmail{
		Email:      inp.Email,
		SecretCode: secretCode,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	return s.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
}

func (s *AuthService) SignIn(ctx *gin.Context, inp domain_auth.SignInInput) (domain_auth.SignInOutput, error) {
	secretCodeHash, err := s.hasher.HashCode(inp.SecretCode)
	if err != nil {
		return domain_auth.SignInOutput{}, err
	}

	verifyEmailParams := repository.GetVerifyEmailParams{
		Email:      inp.Email,
		SecretCode: secretCodeHash,
	}

	verifyEmail, err := s.repoVerifyEmails.Get(ctx, verifyEmailParams)
	if err != nil {
		return domain_auth.SignInOutput{}, err
	}

	if time.Now().After(verifyEmail.ExpiresAt) {
		return domain_auth.SignInOutput{}, domain.ErrSecretCodeExpired
	}

	err = s.repoVerifyEmails.DeleteByID(ctx, verifyEmail.ID)
	if err != nil {
		return domain_auth.SignInOutput{}, err
	}

	user, err := s.repoUsers.GetByEmail(ctx, inp.Email)
	if err != nil {
		return domain_auth.SignInOutput{}, err
	}

	tokens, err := s.createSession(ctx, user.ID)
	if err != nil {
		return domain_auth.SignInOutput{}, err
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
		return domain_auth.SignInOutput{}, err
	}

	return tokens, nil
}

func (s *AuthService) createSession(ctx *gin.Context, id uuid.UUID) (domain_auth.SignInOutput, error) {
	var res domain_auth.SignInOutput

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

	sessionParams := repository.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       id,
		RefreshToken: res.RefreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIP:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	}

	if _, err := s.repoSessions.Create(ctx, sessionParams); err != nil {
		return domain_auth.SignInOutput{}, err
	}

	return res, nil
}

func (s *AuthService) RefreshToken(
	ctx context.Context,
	inp domain_auth.RefreshTokenInput,
) (domain_auth.RefreshTokenOutput, error) {
	var res domain_auth.RefreshTokenOutput

	refreshPayload, err := s.tokenManager.VerifyToken(inp.RefreshToken)
	if err != nil {
		return domain_auth.RefreshTokenOutput{}, err
	}

	session, err := s.repoSessions.Get(ctx, refreshPayload.ID)
	if err != nil {
		return domain_auth.RefreshTokenOutput{}, err
	}

	if session.IsBlocked {
		return domain_auth.RefreshTokenOutput{}, domain.ErrSessionBlocked
	}

	if refreshPayload.UserID != session.UserID {
		return domain_auth.RefreshTokenOutput{}, domain.ErrIncorrectSessionUser
	}

	if inp.RefreshToken != session.RefreshToken {
		return domain_auth.RefreshTokenOutput{}, domain.ErrMismatchedSession
	}

	if time.Now().After(session.ExpiresAt) {
		return domain_auth.RefreshTokenOutput{}, domain.ErrExpiredToken
	}

	accessToken, _, err := s.tokenManager.CreateToken(
		refreshPayload.UserID,
		s.authConfig.JWT.RefreshTokenTTL,
	)
	if err != nil {
		return domain_auth.RefreshTokenOutput{}, err
	}

	res.AccessToken = accessToken

	return res, nil
}
