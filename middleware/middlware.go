package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
	"warehouse/helper"

	"github.com/andybalholm/brotli"
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

// Minimum size to compress (1KB)
const minSize = 1024

// Files to skip
var skipExtensions = []string{
	".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp",
	".zip", ".gz", ".rar", ".iso", ".mp4", ".mp3",
	".pdf", ".woff", ".woff2",
}

func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// If client accepts Brotli
		accept := c.GetHeader("Accept-Encoding")
		useBrotli := strings.Contains(accept, "br")
		useGzip := strings.Contains(accept, "gzip")

		// Replace writer
		writer := &compressedWriter{
			ResponseWriter: c.Writer,
			useBrotli:      useBrotli,
			useGzip:        !useBrotli && useGzip,
		}
		c.Writer = writer

		c.Next()

		// Finish compression
		writer.Close()
	}
}

type compressedWriter struct {
	gin.ResponseWriter
	brWriter  *brotli.Writer
	gzWriter  *gzip.Writer
	useBrotli bool
	useGzip   bool
}

func (w *compressedWriter) Write(data []byte) (int, error) {

	// Skip small responses
	if len(data) < minSize {
		return w.ResponseWriter.Write(data)
	}

	// Choose Brotli
	if w.useBrotli {
		if w.brWriter == nil {
			w.Header().Set("Content-Encoding", "br")
			w.Header().Del("Content-Length")
			w.brWriter = brotli.NewWriter(w.ResponseWriter)
		}
		return w.brWriter.Write(data)
	}

	// Choose Gzip
	if w.useGzip {
		if w.gzWriter == nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			w.gzWriter = gzip.NewWriter(w.ResponseWriter)
		}
		return w.gzWriter.Write(data)
	}

	// No compression
	return w.ResponseWriter.Write(data)
}

func (w *compressedWriter) Close() {
	if w.brWriter != nil {
		w.brWriter.Close()
	}
	if w.gzWriter != nil {
		w.gzWriter.Close()
	}
}
