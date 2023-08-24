package service

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/repository"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/email"
	"github.com/b0shka/backend/pkg/hash"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsersService struct {
	repo         repository.Users
	hasher       hash.Hasher
	tokenManager auth.Manager
	emailService email.EmailService
	emailConfig  config.EmailConfig
	authConfig   config.AuthConfig
}

func NewUsersService(
	repo repository.Users,
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

	emailConfig := domain.VerifyEmailConfig{
		Subject: s.emailConfig.Subjects.Verify,
		Content: content.String(),
	}

	err = s.emailService.SendEmail(emailConfig, email)
	if err != nil {
		return err
	}

	verifyEmail := domain.VerifyEmail{
		Email:          email,
		SecretCodeHash: secretCodeHash,
		ExpiresAt:      time.Now().Add(s.authConfig.SercetCodeLifetime).Unix(),
	}
	return s.repo.AddVerifyEmail(ctx, verifyEmail)
}

func (s *UsersService) SignIn(ctx *gin.Context, inp UserSignInInput) (Tokens, error) {
	secretCodeStr := strconv.Itoa(int(inp.SecretCode))

	secretCodeHash, err := s.hasher.HashCode(secretCodeStr)
	if err != nil {
		return Tokens{}, err
	}

	verifyEmail, err := s.repo.GetVerifyEmail(ctx, inp.Email, secretCodeHash)
	if err != nil {
		return Tokens{}, err
	}

	if time.Now().Unix() > verifyEmail.ExpiresAt {
		return Tokens{}, domain.ErrSecretCodeExpired
	}

	err = s.repo.RemoveVerifyEmail(ctx, verifyEmail.ID)
	if err != nil {
		return Tokens{}, err
	}

	user, err := s.repo.GetUser(ctx, inp.Email)
	var userID primitive.ObjectID
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			userID = primitive.NewObjectID()
			err := s.repo.CreateUser(ctx, domain.User{
				ID:        userID,
				Email:     inp.Email,
				CreatedAt: time.Now().Unix(),
			})
			if err != nil {
				return Tokens{}, err
			}
		} else {
			return Tokens{}, err
		}
	} else {
		userID = user.ID
	}

	return s.createSession(ctx, userID)
}

func (s *UsersService) createSession(ctx *gin.Context, id primitive.ObjectID) (Tokens, error) {
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

	session := domain.Session{
		ID:           refreshPayload.ID,
		UserID:       id,
		RefreshToken: res.RefreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIP:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
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

	if time.Now().Unix() > session.ExpiresAt {
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

func (s *UsersService) Get(ctx context.Context, identifier interface{}) (domain.User, error) {
	return s.repo.GetUser(ctx, identifier)
}

func (s *UsersService) Update(ctx context.Context, id primitive.ObjectID, user domain.UserUpdate) error {
	return s.repo.UpdateUser(ctx, id, user)
}
