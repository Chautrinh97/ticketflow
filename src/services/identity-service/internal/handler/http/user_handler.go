package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/services/identity-service/internal/service"
)

type UserHandler struct {
	users UserService
}

func NewUserHandler(users UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	user, err := h.users.GetByID(c.Request.Context(), httpauth.UserID(c))
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserDTO(user))
}

type updateProfileRequest struct {
	FullName  *string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "Dữ liệu cập nhật không hợp lệ"))
		return
	}
	user, err := h.users.UpdateProfile(c.Request.Context(), httpauth.UserID(c), service.UpdateProfileInput{
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserDTO(user))
}
