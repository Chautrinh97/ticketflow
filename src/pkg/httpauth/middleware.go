// Package httpauth provides the Gin auth middlewares shared by every
// service's own REST handlers: local JWT verification plus the
// RequireRole/RequireOwnership pair specified verbatim (signatures) in
// docs/04-security/authorization.md. Split from pkg/authclaims so services
// that only need JWT signing/parsing (no HTTP layer) don't pull in gin.
package httpauth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/authclaims"
)

const (
	ContextUserID    = "auth_user_id"
	ContextRole      = "auth_role"
	ContextJTI       = "auth_jti"
	ContextExpiresAt = "auth_expires_at"
)

// RequireAuth parses and verifies the Authorization: Bearer <token> header
// locally (signature + expiry only). Revocation (blacklist) and live
// banned-account checks happen once, at the API Gateway, via
// identity.proto's CheckSession RPC — backend services trust the gateway
// for that and only need role/ownership checks here.
func RequireAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Thiếu access token"})
			return
		}
		claims, err := authclaims.Parse(secret, strings.TrimPrefix(header, prefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "Access token không hợp lệ hoặc đã hết hạn"})
			return
		}
		c.Set(ContextUserID, claims.UserID())
		c.Set(ContextRole, claims.Role)
		c.Set(ContextJTI, claims.JTI)
		if claims.ExpiresAt != nil {
			c.Set(ContextExpiresAt, claims.ExpiresAt.Time)
		}
		c.Next()
	}
}

func UserID(c *gin.Context) string { v, _ := c.Get(ContextUserID); s, _ := v.(string); return s }
func Role(c *gin.Context) string   { v, _ := c.Get(ContextRole); s, _ := v.(string); return s }
func JTI(c *gin.Context) string    { v, _ := c.Get(ContextJTI); s, _ := v.(string); return s }
func ExpiresAt(c *gin.Context) time.Time {
	v, _ := c.Get(ContextExpiresAt)
	t, _ := v.(time.Time)
	return t
}

// RequireRole only allows requests whose token role is in roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[Role(c)]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "forbidden", "message": "Không có quyền thực hiện hành động này"})
			return
		}
		c.Next()
	}
}

// RequireOwnership matches a resource's real owner (from resourceLookup,
// which MUST query the resource itself — never infer ownership from
// client-supplied input) against the current user; super_admin always
// bypasses this check, per docs/04-security/authorization.md.
func RequireOwnership(resourceLookup func(c *gin.Context) (ownerID string, err error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if Role(c) == "super_admin" {
			c.Next()
			return
		}
		ownerID, err := resourceLookup(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"code": "not_found", "message": "Không tìm thấy tài nguyên"})
			return
		}
		if ownerID != UserID(c) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "forbidden", "message": "Không có quyền thực hiện hành động này"})
			return
		}
		c.Next()
	}
}
