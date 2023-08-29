package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

var ErrInternalServerError = errors.New("test: internal server error")

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user domain.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser domain.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.ID, gotUser.ID)
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.Photo, gotUser.Photo)
	require.NotEmpty(t, gotUser.CreatedAt)
}

func requireBodyMatchTokens(t *testing.T, body *bytes.Buffer, token service.Tokens) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTokens service.Tokens
	err = json.Unmarshal(data, &gotTokens)

	require.NoError(t, err)
	require.Equal(t, token.AccessToken, gotTokens.AccessToken)
	require.Equal(t, token.RefreshToken, gotTokens.RefreshToken)
}

func requireBodyMatchRefershToken(t *testing.T, body *bytes.Buffer, token service.RefreshToken) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTokens service.RefreshToken
	err = json.Unmarshal(data, &gotTokens)

	require.NoError(t, err)
	require.Equal(t, token.AccessToken, gotTokens.AccessToken)
}

func randomUser(t *testing.T) repository.User {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	user := repository.User{
		ID:       id,
		Email:    utils.RandomEmail(),
		Username: utils.RandomString(10),
		Photo: pgtype.Text{
			String: fmt.Sprintf("https://%s.png", utils.RandomString(7)),
			Valid:  true,
		},
		CreatedAt: time.Now(),
	}

	return user
}

func TestHandler_sendCodeEmail(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, email string)

	tests := []struct {
		name         string
		body         gin.H
		email        string
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			body: gin.H{
				"email": "email@ya.ru",
			},
			email: "email@ya.ru",
			mockBehavior: func(s *mock_service.MockUsers, email string) {
				s.EXPECT().SendCodeEmail(gomock.Any(), email).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "error send code",
			body: gin.H{
				"email": "email@ya.ru",
			},
			mockBehavior: func(s *mock_service.MockUsers, email string) {
				s.EXPECT().SendCodeEmail(gomock.Any(), gomock.Any()).
					Return(ErrInternalServerError)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
		},
		{
			name: "empty fields",
			body: gin.H{
				"email": "",
			},
			mockBehavior: func(s *mock_service.MockUsers, email string) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
		{
			name: "invalid email",
			body: gin.H{
				"email": "email@",
			},
			mockBehavior: func(s *mock_service.MockUsers, email string) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.email)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST("/send-code", handler.sendCodeEmail)

			recorder := httptest.NewRecorder()

			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPost,
				"/send-code",
				bytes.NewReader(data),
			)

			router.ServeHTTP(recorder, req)

			require.Equal(t, testCase.statusCode, recorder.Code)
			require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_userSignIn(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, input domain.UserSignIn)

	tokens := service.Tokens{
		SessionID:             uuid.New(),
		RefreshToken:          utils.RandomString(10),
		RefreshTokenExpiresAt: time.Now().Add(time.Hour * 720),
		AccessToken:           utils.RandomString(10),
		AccessTokenExpiresAt:  time.Now().Add(time.Minute * 15),
	}

	user := repository.User{
		ID:       uuid.New(),
		Email:    utils.RandomEmail(),
		Username: utils.RandomString(10),
		Photo: pgtype.Text{
			String: fmt.Sprintf("https://%s.png", utils.RandomString(7)),
			Valid:  true,
		},
		CreatedAt: time.Now(),
	}

	tests := []struct {
		name          string
		body          gin.H
		userInput     domain.UserSignIn
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			userInput: domain.UserSignIn{
				Email:      "email@ya.ru",
				SecretCode: 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {
				s.EXPECT().
					SignIn(gomock.Any(), input).
					Return(user, tokens, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTokens(t, recorder.Body, tokens)
			},
		},
		{
			name: "error sign in",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(repository.User{}, service.Tokens{}, ErrInternalServerError)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error invalid secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(repository.User{}, service.Tokens{}, domain.ErrSecretCodeInvalid)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrSecretCodeInvalid),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error expired secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(repository.User{}, service.Tokens{}, domain.ErrSecretCodeExpired)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrSecretCodeExpired),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "",
				"secret_code": nil,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Equal(
					t,
					`{"message":"invalid input body"}`,
					recorder.Body.String(),
				)
			},
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": nil,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Equal(
					t,
					`{"message":"invalid input body"}`,
					recorder.Body.String(),
				)
			},
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Equal(
					t,
					`{"message":"invalid input body"}`,
					recorder.Body.String(),
				)
			},
		},
		{
			name: "invalid input secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 12345,
			},
			mockBehavior: func(s *mock_service.MockUsers, input domain.UserSignIn) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Equal(
					t,
					`{"message":"invalid input body"}`,
					recorder.Body.String(),
				)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userInput)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST("/sign-in", handler.userSignIn)

			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost,
				"/sign-in",
				bytes.NewReader(data),
			)

			router.ServeHTTP(recorder, req)
			testCase.checkResponse(recorder)
		})
	}
}

func TestHandler_refreshToken(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, refreshToken string)

	token := service.RefreshToken{
		AccessToken:          utils.RandomString(10),
		AccessTokenExpiresAt: time.Now().Add(time.Minute * 15),
	}

	tests := []struct {
		name          string
		body          gin.H
		refreshToken  string
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), refreshToken).
					Return(token, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchRefershToken(t, recorder.Body, token)
			},
		},
		{
			name: "error refresh token",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, ErrInternalServerError)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error session not found",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrSessionNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrSessionNotFound),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error session blocked",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrSessionBlocked)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrSessionBlocked),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error incorrect session user",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrIncorrectSessionUser)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrIncorrectSessionUser),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error mismatched session",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrMismatchedSession)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrMismatchedSession),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error expires token",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrExpiredToken)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrExpiredToken),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "error invalid token",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			refreshToken: "refresh_token",
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(service.RefreshToken{}, domain.ErrInvalidToken)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrInvalidToken),
					recorder.Body.String(),
				)
			},
		},
		{
			name: "empty fields",
			body: gin.H{
				"refresh_token": "",
			},
			mockBehavior: func(s *mock_service.MockUsers, refreshToken string) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				require.Equal(
					t,
					`{"message":"invalid input body"}`,
					recorder.Body.String(),
				)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.refreshToken)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST("/refresh", handler.refreshToken)

			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost,
				"/refresh",
				bytes.NewReader(data),
			)

			router.ServeHTTP(recorder, req)
			testCase.checkResponse(recorder)
		})
	}
}

func TestHandler_getUserById(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userID uuid.UUID)

	user := randomUser(t)

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupAuth     func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:   "ok",
			userID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, domain.User{
					ID:        user.ID,
					Email:     user.Email,
					Username:  user.Username,
					Photo:     user.Photo.String,
					CreatedAt: user.CreatedAt,
				})
			},
		},
		{
			name:      "no authorization",
			userID:    user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				require.Equal(
					t,
					`{"message":"empty authorization header"}`,
					recorder.Body.String(),
				)
			},
		},
		{
			name:   "user not found",
			userID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(repository.User{}, repository.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, domain.ErrUserNotFound),
					recorder.Body.String(),
				)
			},
		},
		{
			name:   "error get user",
			userID: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().
					GetByID(gomock.Any(), userID).
					Return(repository.User{}, ErrInternalServerError)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
					recorder.Body.String(),
				)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			tokenManager, err := auth.NewJWTManager(utils.RandomString(32))
			require.NoError(t, err)

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userID)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.GET(
				"/",
				userIdentity(tokenManager),
				handler.getUserByID,
			)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			testCase.setupAuth(t, req, tokenManager)
			router.ServeHTTP(recorder, req)

			// require.Equal(t, testCase.statusCode, recorder.Code)
			// require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_updateUser(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate)

	userID, err := uuid.NewRandom()
	require.NoError(t, err)

	tests := []struct {
		name         string
		body         gin.H
		userID       uuid.UUID
		user         domain.UserUpdate
		setupAuth    func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			body: gin.H{
				"username": "vanya",
				"photo":    "",
			},
			userID: userID,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userID, user).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "ok",
			body: gin.H{
				"username": "vanya",
				"photo":    "https://photo.png",
			},
			userID: userID,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "https://photo.png",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userID, user).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "no authorization",
			body: gin.H{
				"username": "vanya",
				"photo":    "https://photo.png",
			},
			userID: userID,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			statusCode:   401,
			responseBody: `{"message":"empty authorization header"}`,
		},
		{
			name: "error update user",
			body: gin.H{
				"username": "vanya",
				"photo":    "",
			},
			userID: userID,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userID, user).Return(ErrInternalServerError)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
		},
		{
			name: "empty fields",
			body: gin.H{
				"username": "",
				"photo":    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			tokenManager, err := auth.NewJWTManager(utils.RandomString(32))
			require.NoError(t, err)

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userID, testCase.user)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST(
				"/update",
				userIdentity(tokenManager),
				handler.updateUser,
			)

			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPost,
				"/update",
				bytes.NewReader(data),
			)

			testCase.setupAuth(t, req, tokenManager)
			router.ServeHTTP(recorder, req)

			require.Equal(t, testCase.statusCode, recorder.Code)
			require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_deleteUser(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userID uuid.UUID)

	userID, err := uuid.NewRandom()
	require.NoError(t, err)

	tests := []struct {
		name         string
		userID       uuid.UUID
		setupAuth    func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name:   "ok",
			userID: userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), userID).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name:      "no authorization",
			userID:    userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
			},
			statusCode:   401,
			responseBody: `{"message":"empty authorization header"}`,
		},
		{
			name:   "user not found",
			userID: userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(repository.ErrRecordNotFound)
			},
			statusCode:   404,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, domain.ErrUserNotFound),
		},
		{
			name:   "error delete user",
			userID: userID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userID uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), userID).Return(ErrInternalServerError)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, ErrInternalServerError),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			tokenManager, err := auth.NewJWTManager(utils.RandomString(32))
			require.NoError(t, err)

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userID)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.GET(
				"/delete",
				userIdentity(tokenManager),
				handler.deleteUser,
			)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				"/delete",
				nil,
			)

			testCase.setupAuth(t, req, tokenManager)
			router.ServeHTTP(recorder, req)

			require.Equal(t, testCase.statusCode, recorder.Code)
			require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}
