package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS restricts cross-origin requests to the single configured frontend
// origin with credentials allowed — required because the refresh-token
// cookie is cross-origin whenever the frontend and gateway run on
// different hosts/ports (docs/06-frontend/README.md).
func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
