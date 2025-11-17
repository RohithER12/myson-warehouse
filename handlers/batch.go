package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var batchRepo = repo.NewBatchRepo()

func CreateBatchHandler(c *gin.Context) {
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
	var batchData models.Batch
	if err := c.ShouldBindJSON(&batchData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	batchData.Status = "active"
	batchData.WarehouseID = warehouseId
	id, err := batchRepo.AddBatch(context.Background(), &batchData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "batch_id": id})
}

// GetAllBatchesHandler fetches all batches
func GetAllBatchesHandler(c *gin.Context) {
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
	batches, err := batchRepo.GetAllBatchesCoreData(context.Background(), warehouseId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batches})
}

// GetBatchByIDHandler fetches batch by ID
func GetBatchByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	batchID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	batch, err := batchRepo.GetBatchCoreDataByID(context.Background(), uint(batchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}
	log.Println("got it here")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": batch})
}

// GetBatchesByProductIDHandler fetches all batches containing a specific product
func GetBatchesByProductIDHandler(c *gin.Context) {
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
	productID := c.Param("id")
	batches, err := batchRepo.GetBatchesByProductID(context.Background(), warehouseId, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	log.Println("got it here")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": batches})
}

// /// ///// /////
