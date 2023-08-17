package handler

import "github.com/gin-gonic/gin"

type response struct {
	Message string `json:"message"`
}

func (h *Handler) newResponse(c *gin.Context, statusCode int, message string) {
	h.logger.Error(message)
	c.AbortWithStatusJSON(statusCode, response{message})
}
