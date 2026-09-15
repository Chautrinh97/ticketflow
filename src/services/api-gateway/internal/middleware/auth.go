// Package middleware holds the API Gateway's cross-cutting Gin middleware:
// the session/ban check and CORS.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/authclaims"
	"ticketflow/services/api-gateway/internal/identityclient"
)

// OptionalSessionCheck runs on every gateway route. When no bearer token is
// present it simply forwards the request — each backend service enforces
// its own auth requirement per route (public vs protected), so the gateway
// must not block genuinely public routes like GET /events or POST /auth/login.
// When a token IS present, it verifies signature+expiry locally, then calls
// identity-service's CheckSession RPC for the one thing only identity-service
// can answer: is this jti blacklisted (logged out) or the account now banned.
// Backend services still locally re-verify the JWT themselves for
// role/ownership checks — this middleware only gates on revocation/ban.
func OptionalSessionCheck(secret []byte, identityCl *identityclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.Next()
			return
		}

		claims, err := authclaims.Parse(secret, strings.TrimPrefix(header, prefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Access token không hợp lệ hoặc đã hết hạn"})
			return
		}

		blacklisted, status, err := identityCl.CheckSession(c.Request.Context(), claims.JTI, claims.UserID())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": "internal_error", "message": "Không thể xác thực phiên đăng nhập, vui lòng thử lại"})
			return
		}
		if blacklisted || status == "banned" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Phiên đăng nhập không còn hiệu lực"})
			return
		}
		c.Next()
	}
}
