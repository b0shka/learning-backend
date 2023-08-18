package handler

import (
	"errors"
	"net/http"
	"strings"

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
	id, err := h.parseAuthHeader(c)
	if err != nil {
		newResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Set(userCtx, id)
}

func (h *Handler) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", errors.New("empty Authorized header")
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid Authorized header")
	}

	if len(headerParts[1]) == 0 {
		return "", errors.New("token is empty")
	}

	return h.tokenManager.Parse(headerParts[1])
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
