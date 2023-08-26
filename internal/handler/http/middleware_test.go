package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/b0shka/backend/pkg/auth"
	"github.com/b0shka/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addAuthorizationHeader(
	t *testing.T,
	request *http.Request,
	tokenManager auth.Manager,
	authorizationType string,
	userId uuid.UUID,
	duration time.Duration,
) {
	token, payload, err := tokenManager.CreateToken(userId, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestHandler_userIdentity(t *testing.T) {
	userId, err := uuid.NewRandom()
	require.NoError(t, err)

	testTable := []struct {
		name         string
		setupAuth    func(t *testing.T, request *http.Request, tokenManager auth.Manager)
		statusCode   int
		responseBody string
	}{
		{
			name: "ok",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, time.Minute)
			},
			statusCode:   200,
			responseBody: "",
		},
		{
			name: "no authorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
			},
			statusCode:   401,
			responseBody: `{"message":"empty authorization header"}`,
		},
		{
			name: "unsupported authorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, "unsupported", userId, time.Minute)
			},
			statusCode:   401,
			responseBody: fmt.Sprintf(`{"message":"unsupported authorization type: %s"}`, "unsupported"),
		},
		{
			name: "invalid authorization format",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, "", userId, time.Minute)
			},
			statusCode:   401,
			responseBody: `{"message":"invalid authorization header format"}`,
		},
		{
			name: "expired token",
			setupAuth: func(t *testing.T, request *http.Request, tokenManager auth.Manager) {
				addAuthorizationHeader(t, request, tokenManager, authorizationTypeBearer, userId, -time.Minute)
			},
			statusCode:   401,
			responseBody: `{"message":"token has expired"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			tokenManager, err := auth.NewPasetoManager(utils.RandomString(32))
			require.NoError(t, err)

			router := gin.Default()

			router.GET(
				"/identity",
				userIdentity(tokenManager),
				func(c *gin.Context) {
					c.Status(http.StatusOK)
				},
			)

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/identity", nil)

			testCase.setupAuth(t, req, tokenManager)
			router.ServeHTTP(recorder, req)

			assert.Equal(t, testCase.statusCode, recorder.Code)
			assert.Equal(t, testCase.responseBody, recorder.Body.String())
		})
	}
}

func TestGetUserPayload(t *testing.T) {
	userId, err := uuid.NewRandom()
	require.NoError(t, err)

	payload, err := auth.NewPayload(userId, time.Minute)
	require.NoError(t, err)

	var getContext = func(payload *auth.Payload) *gin.Context {
		ctx := &gin.Context{}
		ctx.Set(userCtx, payload)
		return ctx
	}

	var getInvalidContext = func() *gin.Context {
		ctx := &gin.Context{}
		ctx.Set(userCtx, "invalid payload")
		return ctx
	}

	tests := []struct {
		name      string
		ctx       *gin.Context
		payload   *auth.Payload
		shouldErr bool
	}{
		{
			name:      "ok",
			ctx:       getContext(payload),
			payload:   payload,
			shouldErr: false,
		},
		{
			name:      "empty user id",
			ctx:       &gin.Context{},
			shouldErr: true,
		},
		{
			name:      "invalid payload",
			ctx:       getInvalidContext(),
			shouldErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			payload, err := getUserPaylaod(testCase.ctx)

			if testCase.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.payload, payload)
		})
	}
}
