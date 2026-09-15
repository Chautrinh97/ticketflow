package http

import (
	"time"

	"ticketflow/services/identity-service/internal/model"
)

// userDTO mirrors the User schema in api-docs/openapi/identity-service.yaml.
type userDTO struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FullName  *string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func toUserDTO(u *model.User) userDTO {
	return userDTO{
		ID:        u.ID,
		Email:     u.Email,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}
