package handlers

import (
	"context"
	"net/http"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var productStockRepo = repo.NewProductStockRepo()

func GetProductStockWithRentHandler(c *gin.Context) {
	stock, err := productStockRepo.GetProductStockWithRent(context.Background())
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
