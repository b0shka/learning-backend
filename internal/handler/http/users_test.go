package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var errInternalServErr = errors.New("test: internal server error")

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
	type mockBehavior func(s *mock_service.MockUsers, input service.UserSignInInput)

	tests := []struct {
		name         string
		body         gin.H
		userInput    service.UserSignInInput
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			userInput: service.UserSignInInput{
				Email:      "email@ya.ru",
				SecretCode: 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), input).
					Return(service.Tokens{
						RefreshToken:          "refresh_token",
						RefreshTokenExpiresAt: time.Now().Add(time.Hour * 720).Unix(),
						AccessToken:           "access_token",
						AccessTokenExpiresAt:  time.Now().Add(time.Minute * 15).Unix(),
					}, nil)
			},
			statusCode:   200,
			responseBody: fmt.Sprintf(`{"refresh_token":"refresh_token","refresh_token_expites_at":%d,"access_token":"access_token","access_token_expires_at":%d}`, time.Now().Add(time.Hour*720).Unix(), time.Now().Add(time.Minute*15).Unix()),
		},
		{
			name: "error sign in",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(service.Tokens{}, errInternalServErr)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
		},
		{
			name: "error invalid secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(service.Tokens{}, domain.ErrSecretCodeInvalid)
			},
			statusCode:   401,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, domain.ErrSecretCodeInvalid),
		},
		{
			name: "error expired secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {
				s.EXPECT().
					SignIn(gomock.Any(), gomock.Any()).
					Return(service.Tokens{}, domain.ErrSecretCodeExpired)
			},
			statusCode:   401,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, domain.ErrSecretCodeExpired),
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "",
				"secret_code": nil,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": nil,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
		{
			name: "empty fields",
			body: gin.H{
				"email":       "",
				"secret_code": 123456,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
		{
			name: "invalid input secret code",
			body: gin.H{
				"email":       "email@ya.ru",
				"secret_code": 12345,
			},
			mockBehavior: func(s *mock_service.MockUsers, input service.UserSignInInput) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
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

			require.Equal(t, testCase.statusCode, recorder.Code)
			require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_getUserById(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userId primitive.ObjectID)

	userId := primitive.NewObjectID()

	tests := []struct {
		name         string
		userId       primitive.ObjectID
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
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID) {
				s.EXPECT().
					Get(gomock.Any(), userId).
					Return(domain.User{
						ID:        userId,
						Email:     "email@ya.ru",
						Photo:     "",
						Name:      "Vanya",
						CreatedAt: time.Now().Unix(),
					}, nil)
			},
			statusCode:   200,
			responseBody: fmt.Sprintf(`{"id":"%s","email":"email@ya.ru","photo":"","name":"Vanya","created_at":%v}`, userId.Hex(), time.Now().Unix()),
		},
		{
			name:      "no authorization",
			userId:    userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID) {
				s.EXPECT().Get(gomock.Any(), gomock.Any()).Times(0)
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
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID) {
				s.EXPECT().
					Get(gomock.Any(), userId).
					Return(domain.User{}, domain.ErrUserNotFound)
			},
			statusCode:   404,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, domain.ErrUserNotFound),
		},
		{
			name:   "error get user",
			userId: userId,
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID) {
				s.EXPECT().
					Get(gomock.Any(), userId).
					Return(domain.User{}, errInternalServErr)
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

			require.Equal(t, testCase.statusCode, recorder.Code)
			require.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_updateUser(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate)

	userId := primitive.NewObjectID()

	tests := []struct {
		name         string
		body         gin.H
		userId       primitive.ObjectID
		user         domain.UserUpdate
		setupAuth    func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			body: gin.H{
				"photo": "",
				"name":  "Vanya",
			},
			userId: userId,
			user: domain.UserUpdate{
				Photo: "",
				Name:  "Vanya",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "ok",
			body: gin.H{
				"photo": "https://photo.png",
				"name":  "Vanya",
			},
			userId: userId,
			user: domain.UserUpdate{
				Photo: "https://photo.png",
				Name:  "Vanya",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "no authorization",
			body: gin.H{
				"photo": "https://photo.png",
				"name":  "Vanya",
			},
			userId: userId,
			user: domain.UserUpdate{
				Photo: "",
				Name:  "Vanya",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			statusCode:   401,
			responseBody: `{"message":"empty authorization header"}`,
		},
		{
			name: "error update user",
			body: gin.H{
				"photo": "",
				"name":  "Vanya",
			},
			userId: userId,
			user: domain.UserUpdate{
				Photo: "",
				Name:  "Vanya",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), userId, user).Return(errInternalServErr)
			},
			statusCode:   500,
			responseBody: fmt.Sprintf(`{"message":"%s"}`, errInternalServErr),
		},
		{
			name: "empty fields",
			body: gin.H{
				"photo": "",
				"name":  "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID, user domain.UserUpdate) {
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
