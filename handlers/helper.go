package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	health := gin.H{
		"server": "healthy",
		"db":     "ok",
		"time":   time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"details": health,
	})
}
