package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b0shka/backend/internal/service"
	mock_service "github.com/b0shka/backend/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestHandler_userIdentity(t *testing.T) {
	// type mockBehavior func(s *auth.TokenManager, token string)

	testTable := []struct {
		name        string
		headerName  string
		headerValue string
		token       string
		// mockBehavior mockBehavior
		statusCode   int
		responseBody string
	}{
		// {
		// 	name:        "ok",
		// 	headerName:  "Authorization",
		// 	headerValue: "Bearer token",
		// 	token:       "token",
		// 	// mockBehavior: func(s auth.TokenManager, token string) {
		// 	// 	s.EXPECT().Parse(token).Return("id", nil).Times(1)
		// 	// },
		// 	statusCode:   200,
		// 	responseBody: "",
		// },
		// {
		// 	name:        "Parse Error",
		// 	headerName:  "Authorization",
		// 	headerValue: "Bearer token",
		// 	token:       "token",
		// 	mockBehavior: func(s *mock_service.MockUsers, token string) {
		// 		r.EXPECT().ParseToken(token).Return(0, errors.New("invalid token"))
		// 	},
		// 	statusCode:   401,
		// 	responseBody: `{"message":"invalid token"}`,
		// },
		{
			name:        "invalid header name",
			headerName:  "",
			headerValue: "Bearer token",
			token:       "token",
			// mockBehavior: func(s *auth.TokenManager, token string) {},
			statusCode:   401,
			responseBody: `{"message":"empty Authorized header"}`,
		},
		{
			name:        "invalid header value",
			headerName:  "Authorization",
			headerValue: "Bearr token",
			token:       "token",
			// mockBehavior: func(s *auth.TokenManager, token string) {},
			statusCode:   401,
			responseBody: `{"message":"invalid Authorized header"}`,
		},
		{
			name:        "empty token",
			headerName:  "Authorization",
			headerValue: "Bearer ",
			token:       "token",
			// mockBehavior: func(s *auth.TokenManager, token string) {},
			statusCode:   401,
			responseBody: `{"message":"token is empty"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			defer mockCtl.Finish()

			usersService := mock_service.NewMockUsers(mockCtl)

			services := &service.Services{Users: usersService}
			handler := Handler{services: services}

			// testCase.mockBehavior(&handler.tokenManager, testCase.token)

			// gin.SetMode(gin.TestMode)
			router := gin.Default()

			router.GET("/identity", handler.userIdentity, func(c *gin.Context) {
				// id, _ := c.Get(userCtx)
				c.Status(http.StatusOK)
			})

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/identity", nil)
			req.Header.Set(testCase.headerName, testCase.headerValue)

			router.ServeHTTP(recorder, req)

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestGetUserId(t *testing.T) {
	userId := primitive.NewObjectID()

	var getContext = func(id primitive.ObjectID) *gin.Context {
		ctx := &gin.Context{}
		ctx.Set(userCtx, id.Hex())
		return ctx
	}

	tests := []struct {
		name       string
		ctx        *gin.Context
		id         primitive.ObjectID
		shouldFail bool
	}{
		{
			name: "ok",
			ctx:  getContext(userId),
			id:   userId,
		},
		{
			name:       "empty user id",
			ctx:        &gin.Context{},
			shouldFail: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			id, err := getUserId(testCase.ctx)

			if testCase.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.id, id)
		})
	}
}
