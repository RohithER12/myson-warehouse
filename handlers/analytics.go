package handlers

import (
	"context"
	"net/http"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var analyticsRepo = repo.NewAnalyticsRepo()

func GetProductAnalytics(c *gin.Context) {
	id := c.Param("id")

	// Convert string to primitive.ObjectID
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "invalid product ID",
		})
		return
	}

	analytics, err := analyticsRepo.GenerateProductAnalytics(context.Background(), productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: analytics})
}

func GetAllProductsAnalytics(c *gin.Context) {
	allAnalytics, err := analyticsRepo.GenerateAllProductsAnalytics(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: allAnalytics})
}
