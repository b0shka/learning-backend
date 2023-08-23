package http

import (
	"errors"
	"net/http"

	"github.com/b0shka/backend/internal/domain"
	"github.com/b0shka/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		authenticating := users.Group("/auth")
		{
			authenticating.POST("/send-code", h.sendCodeEmail)
			authenticating.POST("/sign-in", h.userSignIn)
		}

		authenticated := users.Group("/").Use(userIdentity(h.tokenManager))
		{
			authenticated.GET("/", h.getUserById)
			authenticated.POST("/update", h.updateUser)
		}
	}
}

type userSendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

//	@Summary		User Send Code Email
//	@Tags			auth
//	@Description	send secret code to email user
//	@ModuleID		sendCodeEmail
//	@Accept			json
//	@Produce		json
//	@Param			input	body		userSendCodeRequest	true	"auth info"
//	@Success		201		{string}	string			"ok"
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/auth/send-code [post]
func (h *Handler) sendCodeEmail(c *gin.Context) {
	var inp userSendCodeRequest
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

type userSignInRequest struct {
	Email      string `json:"email" binding:"required,email"`
	SecretCode int32  `json:"secret_code" bson:"secret_code" binding:"required,min=100000"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

//	@Summary		User SignIn
//	@Tags			auth
//	@Description	user sign in
//	@ModuleID		userSignIn
//	@Accept			json
//	@Produce		json
//	@Param			input	body		userSignInRequest	true	"sign in info"
//	@Success		201		{object}	tokenResponse
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/auth/sign-in [post]
func (h *Handler) userSignIn(c *gin.Context) {
	var inp userSignInRequest
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
			newResponse(c, http.StatusUnauthorized, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken: res.AccessToken,
	})
}

//	@Summary		Get User
//  @Security		UsersAuth
//	@Tags			account
//	@Description	get information account
//	@ModuleID		getUserById
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	domain.User
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/ [get]
func (h *Handler) getUserById(c *gin.Context) {
	// id, err := parseIdFromPath(c, "id")
	// if err != nil {
	// 	h.newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
	// 	return
	// }

	userPayload, err := getUserPaylaod(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := h.services.Users.Get(c, userPayload.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			newResponse(c, http.StatusNotFound, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

//	@Summary		Update User
//  @Security		UsersAuth
//	@Tags			account
//	@Description	update user account
//	@ModuleID		updateUser
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UserUpdate	true	"user info"
//	@Success		201		{string}	string			"ok"
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/update [post]
func (h *Handler) updateUser(c *gin.Context) {
	userPayload, err := getUserPaylaod(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var inp domain.UserUpdate
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
