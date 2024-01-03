package http

import (
	"errors"
	"net/http"

	"github.com/b0shka/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) initAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/send-code", h.sendCodeEmail)
		auth.POST("/sign-in", h.signIn)
		auth.POST("/refresh", h.refreshToken)
	}
}

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary		User Send Code Email
// @Tags			auth
// @Description	send secret code to email user
// @ModuleID		sendCodeEmail
// @Accept			json
// @Produce		json
// @Param			input	body		SendCodeRequest	true	"auth info"
// @Success		200		{string}	string			"ok"
// @Failure		400		{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/auth/send-code [post]
func (h *Handler) sendCodeEmail(c *gin.Context) {
	var req SendCodeRequest
	if err := c.BindJSON(&req); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	err := h.services.Auth.SendCodeEmail(c, NewSendCodeEmailInput(req))
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.Status(http.StatusOK)
}

type SignInRequest struct {
	Email      string `json:"email" binding:"required,email"`
	SecretCode string `json:"secret_code" binding:"required,len=6"`
}

type SignInResponse struct {
	SessionID    uuid.UUID `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
}

// @Summary		User SignIn
// @Tags			auth
// @Description	user sign in
// @ModuleID		userSignIn
// @Accept			json
// @Produce		json
// @Param			input	body		SignInRequest	true	"sign in info"
// @Success		200		{object}	SignInResponse
// @Failure		400,401	{object}	response
// @Failure		500		{object}	response
// @Failure		default	{object}	response
// @Router			/users/auth/sign-in [post]
func (h *Handler) signIn(c *gin.Context) {
	var req SignInRequest
	if err := c.BindJSON(&req); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	res, err := h.services.Auth.SignIn(c, NewSignInInput(req))
	if err != nil {
		if errors.Is(err, domain.ErrSecretCodeInvalid) || errors.Is(err, domain.ErrSecretCodeExpired) {
			newResponse(c, http.StatusUnauthorized, err.Error())

			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())

		return
	}

	c.JSON(http.StatusOK, NewSignInResponse(res))
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// @Summary		User Refresh Token
// @Tags			auth
// @Description	user refresh token
// @ModuleID		refreshToken
// @Accept			json
// @Produce		json
// @Param			input	body		RefreshTokenRequest	true	"refresh info"
// @Success		200			{object}	RefreshTokenResponse
// @Failure		400,401,404	{object}	response
// @Failure		500			{object}	response
// @Failure		default		{object}	response
// @Router			/users/auth/refresh [post]
func (h *Handler) refreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.BindJSON(&req); err != nil {
		newResponse(c, http.StatusBadRequest, domain.ErrInvalidInput.Error())

		return
	}

	res, err := h.services.Auth.RefreshToken(c, NewRefreshTokenInput(req))
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

	c.JSON(http.StatusOK, NewRefreshTokenResponse(res))
}
