package http

import (
	"errors"
	"net/http"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/user")
	{
		authenticating := users.Group("/auth")
		{
			authenticating.POST("/send-code", h.sendCodeEmail)
			authenticating.POST("/sign-in", h.userSignIn)
		}

		authenticated := users.Group("/", h.userIdentity)
		{
			authenticated.GET("/", h.getUserById)
			authenticated.POST("/update", h.updateUser)
		}
	}
}

type userEmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

type userSignInInput struct {
	Email      string `json:"email" binding:"required,email"`
	SecretCode int32  `json:"secret_code" bson:"secret_code" binding:"required,min=100000"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) sendCodeEmail(c *gin.Context) {
	var inp userEmailInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
		return
	}

	err := h.services.Users.SendCodeEmail(c, inp.Email)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp userSignInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
		return
	}

	res, err := h.services.Users.SignIn(c, service.UserSignInInput{
		Email:      inp.Email,
		SecretCode: inp.SecretCode,
	})
	if err != nil {
		if errors.Is(err, domain.ErrSecretCodeInvalid) || errors.Is(err, domain.ErrSecretCodeExpired) {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken: res.AccessToken,
	})
}

func (h *Handler) getUserById(c *gin.Context) {
	// id, err := parseIdFromPath(c, "id")
	// if err != nil {
	// 	h.newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
	// 	return
	// }

	id, err := getUserId(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := h.services.Users.Get(c, id)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) updateUser(c *gin.Context) {
	id, err := getUserId(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var inp domain.UserUpdate
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
		return
	}

	inp.ID = id
	if err = h.services.Users.Update(c, inp); err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
