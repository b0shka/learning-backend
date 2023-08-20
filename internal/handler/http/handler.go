package http

import (
	"net/http"

	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/auth"

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

	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := router.Group("/api")
	{
		h.initUsersRoutes(api)
	}

	return router
}

// func parseIdFromPath(c *gin.Context, param string) (primitive.ObjectID, error) {
// 	idParam := c.Param(param)
// 	if idParam == "" {
// 		return primitive.ObjectID{}, errors.New("empty id param")
// 	}

// 	id, err := primitive.ObjectIDFromHex(idParam)
// 	if err != nil {
// 		return primitive.ObjectID{}, errors.New("invalid id param")
// 	}

// 	return id, nil
// }
