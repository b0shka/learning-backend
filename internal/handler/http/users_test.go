package http

import (
	"bytes"
	"database/sql"
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
	"github.com/stretchr/testify/require"
)

var errInternalServErr = errors.New("test: internal server error")

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

func randomUser(t *testing.T) (user repository.User) {
	id, err := uuid.NewRandom()
	require.NoError(t, err)

	user = repository.User{
		ID:        id,
		Email:     utils.RandomEmail(),
		Username:  utils.RandomString(10),
		Photo:     fmt.Sprintf("https://%s", utils.RandomString(7)),
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
					Return(errInternalServErr)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
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
				"POST",
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
		RefreshToken:          utils.RandomString(10),
		RefreshTokenExpiresAt: time.Now().Add(time.Hour * 720),
		AccessToken:           utils.RandomString(10),
		AccessTokenExpiresAt:  time.Now().Add(time.Minute * 15),
	}

	user := repository.User{
		ID:        uuid.New(),
		Email:     utils.RandomEmail(),
		Username:  utils.RandomString(10),
		Photo:     fmt.Sprintf("https://%s", utils.RandomString(7)),
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
					Return(repository.User{}, service.Tokens{}, errInternalServErr)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
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
				"POST",
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
					Return(service.RefreshToken{}, errInternalServErr)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
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
				"POST",
				"/refresh",
				bytes.NewReader(data),
			)

			router.ServeHTTP(recorder, req)
			testCase.checkResponse(recorder)
		})
	}
}

func TestHandler_getUserById(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userId uuid.UUID)

	user := randomUser(t)

	tests := []struct {
		name          string
		userId        uuid.UUID
		setupAuth     func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:   "ok",
			userId: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().
					GetById(gomock.Any(), userId).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, domain.User{
					ID:        user.ID,
					Email:     user.Email,
					Username:  user.Username,
					Photo:     user.Photo,
					CreatedAt: user.CreatedAt,
				})
			},
		},
		{
			name:      "no authorization",
			userId:    user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().GetById(gomock.Any(), gomock.Any()).Times(0)
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
			userId: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().
					GetById(gomock.Any(), userId).
					Return(repository.User{}, sql.ErrNoRows)
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
			userId: user.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, user.ID, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().
					GetById(gomock.Any(), userId).
					Return(repository.User{}, errInternalServErr)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				require.Equal(
					t,
					fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
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
			testCase.mockBehavior(usersService, testCase.userId)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.GET(
				"/",
				userIdentity(tokenManager),
				handler.getUserById,
			)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				"GET",
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
	type mockBehavior func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate)

	userId, err := uuid.NewRandom()
	require.NoError(t, err)

	tests := []struct {
		name         string
		body         gin.H
		userId       uuid.UUID
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
			userId: userId,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(nil)
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
			userId: userId,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "https://photo.png",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(nil)
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
			userId: userId,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate) {
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
			userId: userId,
			user: domain.UserUpdate{
				Username: "vanya",
				Photo:    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(errInternalServErr)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
		},
		{
			name: "empty fields",
			body: gin.H{
				"username": "",
				"photo":    "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID, user domain.UserUpdate) {
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
			testCase.mockBehavior(usersService, testCase.userId, testCase.user)

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
				"POST",
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
	type mockBehavior func(s *mock_service.MockUsers, userId uuid.UUID)

	userId, err := uuid.NewRandom()
	require.NoError(t, err)

	tests := []struct {
		name         string
		userId       uuid.UUID
		setupAuth    func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name:   "ok",
			userId: userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), userId).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name:      "no authorization",
			userId:    userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
			},
			statusCode:   401,
			responseBody: `{"message":"empty authorization header"}`,
		},
		{
			name:   "user not found",
			userId: userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(sql.ErrNoRows)
			},
			statusCode:   404,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, domain.ErrUserNotFound),
		},
		{
			name:   "error delete user",
			userId: userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId uuid.UUID) {
				s.EXPECT().Delete(gomock.Any(), userId).Return(errInternalServErr)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			tokenManager, err := auth.NewJWTManager(utils.RandomString(32))
			require.NoError(t, err)

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userId)

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
				"GET",
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
