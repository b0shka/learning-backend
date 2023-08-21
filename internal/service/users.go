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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsersService struct {
	repo         repository.Users
	hasher       hash.PasswordHasher
	tokenManager auth.TokenManager
	emailService email.EmailService
	emailConfig  config.EmailConfig
	authConfig   config.AuthConfig
}

func NewUsersService(
	repo repository.Users,
	hasher hash.PasswordHasher,
	tokenManager auth.TokenManager,
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
		ExpiredAt:      time.Now().Unix() + int64(s.authConfig.SercetCodeLifetime),
	}
	return s.repo.AddVerifyEmail(ctx, verifyEmail)
}

func (s *UsersService) SignIn(ctx context.Context, inp UserSignInInput) (Tokens, error) {
	secretCodeStr := strconv.Itoa(int(inp.SecretCode))

	secretCodeHash, err := s.hasher.HashCode(secretCodeStr)
	if err != nil {
		return Tokens{}, err
	}

	verifyEmail, err := s.repo.GetVerifyEmail(ctx, inp.Email, secretCodeHash)
	if err != nil {
		return Tokens{}, err
	}

	if time.Now().Unix() > verifyEmail.ExpiredAt {
		return Tokens{}, domain.ErrSecretCodeExpired
	}

	err = s.repo.RemoveVerifyEmail(ctx, verifyEmail.ID)
	if err != nil {
		return Tokens{}, err
	}

	user, err := s.repo.Get(ctx, inp.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			err := s.repo.Create(ctx, domain.User{
				Email:     inp.Email,
				CreatedAt: time.Now().Unix(),
			})
			if err != nil {
				return Tokens{}, err
			}
		}
		return Tokens{}, err
	}

	return s.createSession(user.ID)
}

func (s *UsersService) createSession(id primitive.ObjectID) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.NewJWT(
		id.Hex(),
		s.authConfig.JWT.AccessTokenTTL,
	)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *UsersService) Get(ctx context.Context, identifier interface{}) (domain.User, error) {
	return s.repo.Get(ctx, identifier)
}

func (s *UsersService) Update(ctx context.Context, user domain.UserUpdate) error {
	return s.repo.Update(ctx, user)
}
