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
	warehouseIdAny, exists := c.Get("warehouse_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "warehouse_id not found in token"})
		return
	}
	warehouseId, ok := warehouseIdAny.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "invalid warehouse_id type"})
		return
	}
	data, err := analyticsRepo.GetAnalytics(context.Background(), warehouseId, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}
