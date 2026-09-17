package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/services/identity-service/internal/service"
)

// refreshCookieName per docs/04-security/authentication.md: httpOnly +
// Secure + SameSite=Strict, never exposed in a response body.
const refreshCookieName = "refresh_token"

type AuthHandler struct {
	auth         AuthService
	cookieSecure bool
}

func NewAuthHandler(auth AuthService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{auth: auth, cookieSecure: cookieSecure}
}

type loginRequest struct {
	FirebaseIDToken string `json:"firebase_id_token" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "firebase_id_token là bắt buộc"))
		return
	}
	result, err := h.auth.Login(c.Request.Context(), req.FirebaseIDToken)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
		"user":         toUserDTO(result.User),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, _ := c.Cookie(refreshCookieName)
	result, err := h.auth.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	c.JSON(http.StatusOK, gin.H{
		"access_token": result.AccessToken,
		"expires_in":   result.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(refreshCookieName)
	if err := h.auth.Logout(c.Request.Context(), httpauth.JTI(c), httpauth.ExpiresAt(c), refreshToken); err != nil {
		apperr.Respond(c, err)
		return
	}
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, token, int(service.RefreshTokenTTL.Seconds()), "/", "", h.cookieSecure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(refreshCookieName, "", -1, "/", "", h.cookieSecure, true)
}
