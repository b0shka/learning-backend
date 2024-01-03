package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users").Use(userIdentity(h.tokenManager))
	{
		users.GET("/", h.getUserByID)
		users.DELETE("/", h.deleteUser)
	}
}

type GetUserResponse struct {
	ID        uuid.UUID `json:"id" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}

// @Summary		Get User
// @Security		UsersAuth
// @Tags			account
// @Description	get information account
// @ModuleID		getUserById
// @Accept			json
// @Produce		json
// @Success		200		{object}	GetUserResponse
// @Failure		400,404	{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/ [get]
func (h *Handler) getUserByID(c *gin.Context) {
	userPayload, err := getUserPayload(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	// id, err := parseIdFromPath(c, "id")
	// if err != nil {
	// 	newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
	// 	return
	// }

	user, err := h.services.Users.GetByID(c, userPayload.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			newResponse(c, http.StatusNotFound, domain.ErrUserNotFound.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, NewGetUserResponse(user))
}

// @Summary		Delete User
// @Security		UsersAuth
// @Tags			account
// @Description	delete user account
// @ModuleID		deleteUser
// @Accept			json
// @Produce		json
// @Success		200		{string}	string			"ok"
// @Failure		400,404	{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/ [delete]
func (h *Handler) deleteUser(c *gin.Context) {
	userPayload, err := getUserPayload(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	if err = h.services.Users.Delete(c, userPayload.UserID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			newResponse(c, http.StatusNotFound, domain.ErrUserNotFound.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.Status(http.StatusOK)
}
