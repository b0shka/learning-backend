package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
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
					Return(service.Tokens{AccessToken: "access_token"}, nil)
			},
			statusCode:   200,
			responseBody: `{"access_token":"access_token"}`,
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
			statusCode:   400,
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
			statusCode:   400,
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

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_getUserById(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, userId primitive.ObjectID)

	userId := primitive.NewObjectID()

	tests := []struct {
		name         string
		userId       primitive.ObjectID
		mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		{
			name:   "ok",
			userId: userId,
			mockBehavior: func(s *mock_service.MockUsers, userId primitive.ObjectID) {
				s.EXPECT().
					Get(gomock.Any(), userId).
					Return(domain.User{
						ID:        userId,
						Email:     "email@ya.ru",
						Photo:     "",
						Name:      "Vanya",
						CreatedAt: 1692272560,
					}, nil)
			},
			statusCode:   200,
			responseBody: fmt.Sprintf(`{"id":"%s","email":"email@ya.ru","photo":"","name":"Vanya","created_at":1692272560}`, userId.Hex()),
		},
		{
			name:   "error get user",
			userId: userId,
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

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.userId)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.GET("/", func(c *gin.Context) {
				c.Set(userCtx, testCase.userId.Hex())
			}, handler.getUserById)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				"GET",
				"/",
				nil,
			)

			router.ServeHTTP(recorder, req)

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestHandler_updateUser(t *testing.T) {
	type mockBehavior func(s *mock_service.MockUsers, user domain.UserUpdate)

	userId := primitive.NewObjectID()

	tests := []struct {
		name         string
		body         gin.H
		user         domain.UserUpdate
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
			user: domain.UserUpdate{
				ID:    userId,
				Photo: "",
				Name:  "Vanya",
			},
			mockBehavior: func(s *mock_service.MockUsers, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), user).Return(nil)
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
			user: domain.UserUpdate{
				ID:    userId,
				Photo: "https://photo.png",
				Name:  "Vanya",
			},
			mockBehavior: func(s *mock_service.MockUsers, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), user).Return(nil)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "error update user",
			body: gin.H{
				"photo": "",
				"name":  "Vanya",
			},
			user: domain.UserUpdate{
				ID:    userId,
				Photo: "",
				Name:  "Vanya",
			},
			mockBehavior: func(s *mock_service.MockUsers, user domain.UserUpdate) {
				s.EXPECT().Update(gomock.Any(), user).Return(errInternalServErr)
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
			mockBehavior: func(s *mock_service.MockUsers, user domain.UserUpdate) {},
			statusCode:   400,
			responseBody: `{"message":"invalid input body"}`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			usersService := mock_service.NewMockUsers(mockCtl)
			testCase.mockBehavior(usersService, testCase.user)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			router := gin.Default()
			router.POST("/update", func(c *gin.Context) {
				c.Set(userCtx, testCase.user.ID.Hex())
			}, handler.updateUser)

			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(
				"POST",
				"/update",
				bytes.NewReader(data),
			)

			router.ServeHTTP(recorder, req)

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}
