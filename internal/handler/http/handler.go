package http

import (
	"fmt"
	"net/http"

	"github.com/b0shka/backend/docs"
	"github.com/b0shka/backend/internal/config"
	"github.com/b0shka/backend/internal/service"
	"github.com/b0shka/backend/pkg/auth"

	_ "github.com/b0shka/backend/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

func (h *Handler) InitRoutes(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		corsMiddleware,
	)

	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
