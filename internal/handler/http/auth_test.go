package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b0shka/backend/internal/domain"
	domain_auth "github.com/b0shka/backend/internal/domain/auth"
	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/b0shka/backend/pkg/logger"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var ErrInternalServerError = errors.New("test: internal server error")

func requireBodyMatchTokens(t *testing.T, body *bytes.Buffer, token domain_auth.SignInOutput) {
	var gotTokens domain_auth.SignInOutput
	err := json.Unmarshal(body.Bytes(), &gotTokens)
	logger.Info(gotTokens)

	require.NoError(t, err)
	require.Equal(t, token.AccessToken, gotTokens.AccessToken)
	require.Equal(t, token.RefreshToken, gotTokens.RefreshToken)
}

func requireBodyMatchRefreshToken(t *testing.T, body *bytes.Buffer, token domain_auth.RefreshTokenOutput) {
	var gotTokens domain_auth.RefreshTokenOutput
	err := json.Unmarshal(body.Bytes(), &gotTokens)

	require.NoError(t, err)
	require.Equal(t, token.AccessToken, gotTokens.AccessToken)
}

func TestHandler_sendCodeEmail(t *testing.T) {
	type mockBehavior func(s *mock_service.MockAuth, input domain_auth.SendCodeEmailInput)

	tests := []struct {
		name         string
		body         gin.H
		userInput    domain_auth.SendCodeEmailInput
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			body: gin.H{
				"email": "email@ya.ru",
			},
			userInput: domain_auth.SendCodeEmailInput{
				Email: "email@ya.ru",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SendCodeEmailInput) {
				s.EXPECT().SendCodeEmail(gomock.Any(), input).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "error send code",
			body: gin.H{
				"email": "email@ya.ru",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SendCodeEmailInput) {
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
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SendCodeEmailInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
		{
			name: "invalid email",
			body: gin.H{
				"email": "email@",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SendCodeEmailInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			authService := mock_service.NewMockAuth(mockCtl)
			testCase.mockBehavior(authService, testCase.userInput)

			services := &service.Services{Auth: authService}
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
	type mockBehavior func(s *mock_service.MockAuth, input domain_auth.SignInInput)

	refreshToken, err := utils.RandomString(10)
	require.NoError(t, err)
	accessToken, err := utils.RandomString(10)
	require.NoError(t, err)

	tokens := domain_auth.SignInOutput{
		SessionID:    uuid.New(),
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}

	tests := []struct {
		name          string
		body          gin.H
		userInput     domain_auth.SignInInput
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": "123456",
			},
			userInput: domain_auth.SignInInput{
				Email:      "email@ya.ru",
				SecretCode: "123456",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), input).
					Return(tokens, nil)
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
				"secret_code": "123456",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(domain_auth.SignInOutput{}, ErrInternalServerError)
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
				"secret_code": "123456",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(domain_auth.SignInOutput{}, domain.ErrSecretCodeInvalid)
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
				"secret_code": "123456",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(domain_auth.SignInOutput{}, domain.ErrSecretCodeExpired)
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
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {},
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
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {},
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
				"secret_code": "123456",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {},
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
			name: "small length input secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": "12345",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {},
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
			name: "large length input secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": "1234567",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.SignInInput) {},
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

			authService := mock_service.NewMockAuth(mockCtl)
			testCase.mockBehavior(authService, testCase.userInput)

			services := &service.Services{Auth: authService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST("/sign-in", handler.signIn)

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
	type mockBehavior func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput)

	accessToken, err := utils.RandomString(10)
	require.NoError(t, err)

	token := domain_auth.RefreshTokenOutput{
		AccessToken: accessToken,
	}

	tests := []struct {
		name          string
		body          gin.H
		userInput     domain_auth.RefreshTokenInput
		mockBehavior  mockBehavior
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), input).
					Return(token, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchRefreshToken(t, recorder.Body, token)
			},
		},
		{
			name: "error refresh token",
			body: gin.H{
				"refresh_token": "refresh_token",
			},
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, ErrInternalServerError)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrSessionNotFound)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrSessionBlocked)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrIncorrectSessionUser)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrMismatchedSession)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrExpiredToken)
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
			userInput: domain_auth.RefreshTokenInput{
				RefreshToken: "refresh_token",
			},
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {
				s.EXPECT().
					RefreshToken(gomock.Any(), gomock.Any()).
					Return(domain_auth.RefreshTokenOutput{}, domain.ErrInvalidToken)
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
			mockBehavior: func(s *mock_service.MockAuth, input domain_auth.RefreshTokenInput) {},
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

			authService := mock_service.NewMockAuth(mockCtl)
			testCase.mockBehavior(authService, testCase.userInput)

			services := &service.Services{Auth: authService}
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
