package http

import (
	"errors"
	"net/http"

	"github.com/b0shka/backend/internal/domain"
	repository "github.com/b0shka/backend/internal/repository/postgresql/sqlc"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users").Use(userIdentity(h.tokenManager))
	{
		users.GET("/", h.getUserByID)
		users.PATCH("/", h.updateUser)
		users.DELETE("/", h.deleteUser)
	}
}

// @Summary		Get User
// @Security		UsersAuth
// @Tags			account
// @Description	get information account
// @ModuleID		getUserById
// @Accept			json
// @Produce		json
// @Success		200		{object}	domain.User
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

	c.JSON(http.StatusOK, domain.GetUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Photo:     user.Photo.String,
		CreatedAt: user.CreatedAt,
	})
}

// @Summary		Update User
// @Security		UsersAuth
// @Tags			account
// @Description	update user account
// @ModuleID		updateUser
// @Accept			json
// @Produce		json
// @Param			input	body		domain.UserUpdate	true	"user update info"
// @Success		200		{string}	string			"ok"
// @Failure		400		{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/ [put]
func (h *Handler) updateUser(c *gin.Context) {
	userPayload, err := getUserPayload(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	var inp domain.UpdateUserRequest
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	if err = h.services.Users.Update(c, userPayload.UserID, inp); err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.Status(http.StatusOK)
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
