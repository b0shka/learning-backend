package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "Bearer"

	userCtx = "userCtx"
)

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != http.MethodOptions {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

func userIdentity(tokenManager auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := parseAuthHeader(c, tokenManager)
		if err != nil {
			newResponse(c, http.StatusUnauthorized, err.Error())

			return
		}

		c.Set(userCtx, payload)
	}
}

func parseAuthHeader(c *gin.Context, tokenManager auth.Manager) (*auth.Payload, error) {
	authorizationHeader := c.GetHeader(authorizationHeaderKey)
	if len(authorizationHeader) == 0 {
		return nil, domain.ErrEmptyAuthHeader
	}

	headerParts := strings.Fields(authorizationHeader)
	if len(headerParts) < 2 {
		return nil, domain.ErrInvalidAuthHeaderFormat
	}

	authorizationType := headerParts[0]
	if authorizationType != authorizationTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authorizationType)
	}

	return tokenManager.VerifyToken(headerParts[1])
}

func getUserPaylaod(c *gin.Context) (*auth.Payload, error) {
	return getPayloadByContext(c, userCtx)
}

func getPayloadByContext(c *gin.Context, context string) (*auth.Payload, error) {
	payloadFromCtx, ok := c.Get(context)
	if !ok {
		return nil, fmt.Errorf("%s not found", context)
	}

	payload, ok := payloadFromCtx.(*auth.Payload)
	if !ok {
		return nil, fmt.Errorf("%s is of invalid type", context)
	}

	return payload, nil
}
