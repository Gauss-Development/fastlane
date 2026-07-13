// internal/middleware/role.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"api-gateway/pkg/utils"
)

// RequireRole gates a route to callers whose role is in the allowed set.
// Assumes AuthMiddleware ran first (userRole is set from the validated token).
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("userRole")
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		utils.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Insufficient role")
		c.Abort()
	}
}
