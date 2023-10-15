package http

import (
	"errors"
	"net/http"

	"github.com/b0shka/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/send-code", h.sendCodeEmail)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/refresh", h.refreshToken)
	}
}

// @Summary		User Send Code Email
// @Tags			auth
// @Description	send secret code to email user
// @ModuleID		sendCodeEmail
// @Accept			json
// @Produce		json
// @Param			input	body		domain.SendCodeRequest	true	"auth info"
// @Success		200		{string}	string			"ok"
// @Failure		400		{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/auth/send-code [post]
func (h *Handler) sendCodeEmail(c *gin.Context) {
	var inp domain.SendCodeRequest
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	err := h.services.Auth.SendCodeEmail(c, inp.Email)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.Status(http.StatusOK)
}

// @Summary		User SignIn
// @Tags			auth
// @Description	user sign in
// @ModuleID		userSignIn
// @Accept			json
// @Produce		json
// @Param			input	body		domain.UserSignIn	true	"sign in info"
// @Success		200		{object}	domain.SignInResponse
// @Failure		400,401	{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var inp domain.SignInRequest
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	res, err := h.services.Auth.SignIn(c, inp)
	if err != nil {
		if errors.Is(err, domain.ErrSecretCodeInvalid) || errors.Is(err, domain.ErrSecretCodeExpired) {
			newResponse(c, http.StatusUnauthorized, err.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, domain.SignInResponse{
		SessionID:    res.SessionID,
		RefreshToken: res.RefreshToken,
		AccessToken:  res.AccessToken,
	})
}

// @Summary		User Refresh Token
// @Tags			auth
// @Description	user refresh token
// @ModuleID		refreshToken
// @Accept			json
// @Produce		json
// @Param			input	body		domain.RefreshTokenRequest	true	"refresh info"
// @Success		200			{object}	domain.RefreshTokenResponse
// @Failure		400,401,404	{object}	response
// @Failure		500			{object}	response
// @Failure		default		{object}	response
// @Router			/users/auth/refresh [post]
func (h *Handler) refreshToken(c *gin.Context) {
	var inp domain.RefreshTokenRequest
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	res, err := h.services.Auth.RefreshToken(c, inp.RefreshToken)
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

	c.JSON(http.StatusOK, domain.RefreshTokenResponse{
		AccessToken: res.AccessToken,
	})
}
