package handlers

import (
	"context"
	"net/http"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var stockRepo = repo.NewStockRepo()

// CreateBatchHandler creates a new batch
func CreateBatchHandler(c *gin.Context) {
	var batch models.Batch
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	id, err := stockRepo.AddBatch(context.Background(), &batch)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "batch_id": id})
}

// OffboardProductHandler offboards a product quantity from batches
func OffboardProductHandler(c *gin.Context) {
	var req struct {
		ProductID   string `json:"product_id"`
		OffboardQty int    `json:"offboard_quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := stockRepo.Offboard(context.Background(), req.ProductID, req.OffboardQty); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product offboarded successfully"})
}

// GetAllBatchesHandler fetches all batches
func GetAllBatchesHandler(c *gin.Context) {
	batches, err := stockRepo.GetAllBatches(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batches})
}

// GetBatchByIDHandler fetches batch by ID
func GetBatchByIDHandler(c *gin.Context) {
	batchID := c.Param("id")
	batch, err := stockRepo.GetBatchByID(context.Background(), batchID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batch})
}

// GetBatchesByProductIDHandler fetches all batches containing a specific product
func GetBatchesByProductIDHandler(c *gin.Context) {
	productID := c.Param("id")
	batches, err := stockRepo.GetBatchesByProductID(context.Background(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batches})
}
