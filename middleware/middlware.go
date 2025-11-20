package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"warehouse/helper"

	"github.com/gin-gonic/gin"
)

func AuthJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "missing or invalid Authorization header"})
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := helper.ParseJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid token"})
			return
		}
		// stash claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("warehouse_id", claims.WarehouseId)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRoles(allowed ...string) gin.HandlerFunc {
	allowedSet := map[string]struct{}{}
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleIF, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "role not found in token"})
			return
		}
		role := roleIF.(string)
		if _, ok := allowedSet[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	gw *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.gw.Write(data)
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Only gzip if client supports it
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Set response headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Writer = &gzipWriter{ResponseWriter: c.Writer, gw: gz}

		c.Next()
	}
}
