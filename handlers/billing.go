package handlers

import (
	"context"
	"net/http"
	"strconv"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var billingRepo = repo.NewBillingRepo()

func CreateBillingWithBatchId(c *gin.Context) {
	var input models.BillingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	billing, err := billingRepo.CreateBillingWithBatchId(context.Background(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Billing created successfully", "data": billing})
}
func CreateBillingWithOutBatchId(c *gin.Context) {
	var input models.BillingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	billing, err := billingRepo.CreateBillingWithOutBatchId(context.Background(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Billing created successfully", "data": billing})
}

func GetBillByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	batchID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	batch, err := billingRepo.GetByID(context.Background(), uint(batchID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batch})
}

func GetAllBillsHandler(c *gin.Context) {

	batches, err := billingRepo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": batches})
}
