package handlers

import (
	"context"
	"net/http"
	"strconv"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var productStockRepo = repo.NewProductStockRepo()

func GetProductStockWithRentHandler(c *gin.Context) {
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
	stock, err := productStockRepo.GetProductStockWithRent(context.Background(), warehouseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stock,
	})
}

func GetAllProductStockHandler(c *gin.Context) {
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
	stock, err := productStockRepo.GetAllproducts(context.Background(), warehouseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stock,
	})
}

func SearchStockProductData(c *gin.Context) {

	idStr := c.Param("product_id")
	productId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	stock, err := productStockRepo.GetStockProductData(context.Background(), uint(productId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stock,
	})
}
