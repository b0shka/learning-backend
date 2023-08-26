package http

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/b0shka/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		auth := users.Group("/auth")
		{
			auth.POST("/send-code", h.sendCodeEmail)
			auth.POST("/sign-in", h.userSignIn)
			auth.POST("/refresh", h.refreshToken)
		}

		account := users.Group("/").Use(userIdentity(h.tokenManager))
		{
			account.GET("/", h.getUserById)
			account.POST("/update", h.updateUser)
			account.GET("/delete", h.deleteUser)
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

type userSignInResponse struct {
	RefreshToken          string      `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time   `json:"refresh_token_expites_at"`
	AccessToken           string      `json:"access_token"`
	AccessTokenExpiresAt  time.Time   `json:"access_token_expires_at"`
	User                  domain.User `json:"user"`
}

//	@Summary		User SignIn
//	@Tags			auth
//	@Description	user sign in
//	@ModuleID		userSignIn
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UserSignIn	true	"sign in info"
//	@Success		201		{object}	tokenResponse
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/auth/sign-in [post]
func (h *Handler) userSignIn(c *gin.Context) {
	var inp domain.UserSignIn
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
		return
	}

	user, res, err := h.services.Users.SignIn(c, inp)
	if err != nil {
		if errors.Is(err, domain.ErrSecretCodeInvalid) || errors.Is(err, domain.ErrSecretCodeExpired) {
			newResponse(c, http.StatusUnauthorized, err.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, userSignInResponse{
		RefreshToken:          res.RefreshToken,
		RefreshTokenExpiresAt: res.RefreshTokenExpiresAt,
		AccessToken:           res.AccessToken,
		AccessTokenExpiresAt:  res.AccessTokenExpiresAt,
		User: domain.User{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Photo:     user.Photo,
			CreatedAt: user.CreatedAt,
		},
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

//	@Summary		User Refresh Token
//	@Tags			auth
//	@Description	user refresh token
//	@ModuleID		refreshToken
//	@Accept			json
//	@Produce		json
//	@Param			input	body		refreshTokenRequest	true	"refresh info"
//	@Success		201		{object}	refreshTokenResponse
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/auth/refresh [post]
func (h *Handler) refreshToken(c *gin.Context) {
	var inp refreshTokenRequest
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())
		return
	}

	res, err := h.services.Users.RefreshToken(c, inp.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			newResponse(c, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, domain.ErrSessionBlocked) ||
			errors.Is(err, domain.ErrIncorrectSessionUser) ||
			errors.Is(err, domain.ErrMismatchedSession) ||
			errors.Is(err, domain.ErrExpiredToken) ||
			errors.Is(err, domain.ErrInvalidToken) {
			newResponse(c, http.StatusUnauthorized, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, refreshTokenResponse{
		AccessToken:          res.AccessToken,
		AccessTokenExpiresAt: res.AccessTokenExpiresAt,
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

	user, err := h.services.Users.GetById(c, userPayload.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newResponse(c, http.StatusNotFound, domain.ErrUserNotFound.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, domain.User{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Photo:     user.Photo,
		CreatedAt: user.CreatedAt,
	})
}

//	@Summary		Update User
//  @Security		UsersAuth
//	@Tags			account
//	@Description	update user account
//	@ModuleID		updateUser
//	@Accept			json
//	@Produce		json
//	@Param			input	body		domain.UserUpdate	true	"user update info"
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

//	@Summary		Delete User
//  @Security		UsersAuth
//	@Tags			account
//	@Description	delete user account
//	@ModuleID		deleteUser
//	@Accept			json
//	@Produce		json
//	@Success		201		{string}	string			"ok"
//	@Failure		400,404	{object}	response
//	@Failure		500		{object}	response
//	@Failure		default	{object}	response
//	@Router			/user/delete [get]
func (h *Handler) deleteUser(c *gin.Context) {
	userPayload, err := getUserPaylaod(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err = h.services.Users.Delete(c, userPayload.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newResponse(c, http.StatusNotFound, domain.ErrUserNotFound.Error())
			return
		}
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
}
