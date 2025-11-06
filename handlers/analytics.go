package handlers

import (
	"context"
	"net/http"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var analyticsRepo = repo.NewAnalyticsRepo()

func GetAnalyticsHandler(c *gin.Context) {
	duration := c.Param("duration")

	data, err := analyticsRepo.GetAnalytics(context.Background(), duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
