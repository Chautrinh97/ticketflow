package http

import (
	"github.com/gin-gonic/gin"

	"ticketflow/pkg/httpauth"
)

// NewRouter mounts every Phase 1 endpoint from api-docs/openapi/identity-service.yaml
// under /api/v1, matching that file's `servers: [{url: /api/v1}]` so the
// API Gateway can reverse-proxy paths unchanged.
func NewRouter(jwtSecret []byte, authHandler *AuthHandler, userHandler *UserHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", httpauth.RequireAuth(jwtSecret), authHandler.Logout)

		users := v1.Group("/users")
		users.Use(httpauth.RequireAuth(jwtSecret))
		users.GET("/me", userHandler.GetMe)
		users.PATCH("/me", userHandler.UpdateMe)
	}
	return r
}
