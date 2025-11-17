package handlers

import (
	"context"
	"net/http"
	"strconv"
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
func GetFastAndSlowMovingProductAnalytics(c *gin.Context) {
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
	data, err := analyticsRepo.GetFastAndSlowMovingProductAnalytics(context.Background(), warehouseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func GetProductAnalyticsByIdHandler(c *gin.Context) {
	productIDStr := c.Param("product_id") // assuming path like /analytics/product/:product_id

	if productIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "product_id is required"})
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid product_id"})
		return
	}

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

	data, err := analyticsRepo.GetProductAnalyticsById(c.Request.Context(), warehouseId, uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
