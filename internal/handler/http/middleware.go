package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/b0shka/backend/pkg/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "userId"
)

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

func (h *Handler) userIdentity(c *gin.Context) {
	payload, err := h.parseAuthHeader(c)
	if err != nil {
		newResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	c.Set(userCtx, payload.UserID)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (*auth.Payload, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return nil, errors.New("empty Authorized header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("invalid Authorized header")
	}

	if len(headerParts[1]) == 0 {
		return nil, errors.New("token is empty")
	}

	return h.tokenManager.VerifyToken(headerParts[1])
}

func getUserId(c *gin.Context) (primitive.ObjectID, error) {
	return getIdByContext(c, userCtx)
}

func getIdByContext(c *gin.Context, context string) (primitive.ObjectID, error) {
	idFromCtx, ok := c.Get(context)
	if !ok {
		return primitive.ObjectID{}, errors.New("studentCtx not found")
	}

	idStr, ok := idFromCtx.(string)
	if !ok {
		return primitive.ObjectID{}, errors.New("studentCtx is of invalid type")
	}

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	return id, nil
}
