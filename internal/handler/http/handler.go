package handler

import (
	"errors"

	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/auth"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewHandler(
	services *service.Services,
	tokenManager auth.TokenManager,
) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		corsMiddleware,
	)

	api := router.Group("/api")
	{
		h.initUsersRoutes(api)
	}

	return router
}

func parseIdFromPath(c *gin.Context, param string) (primitive.ObjectID, error) {
	idParam := c.Param(param)
	if idParam == "" {
		return primitive.ObjectID{}, errors.New("empty id param")
	}

	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		return primitive.ObjectID{}, errors.New("invalid id param")
	}

	return id, nil
}
