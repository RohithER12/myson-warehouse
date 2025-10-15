package handlers

import (
	"context"
	"net/http"
	"time"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var billingRepo = repo.NewBillingRepo()

func GenerateBilling(c *gin.Context) {
	var body struct {
		ProductID   string         `json:"product_id"`
		OffboardQty int            `json:"offboard_quantity"`
		StartDate   time.Time      `json:"start_date"`
		EndDate     time.Time      `json:"end_date"`
		RentPerUnit float64        `json:"rent_per_unit"`
		Expenses    []repo.Expense `json:"expenses"`
	}

	// Bind JSON request
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Validate dates
	if body.EndDate.Before(body.StartDate) {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "end_date must be after start_date"})
		return
	}

	// Call repo to generate billing
	bill, err := billingRepo.GenerateBilling(
		context.Background(),
		body.ProductID,
		body.OffboardQty,
		body.StartDate,
		body.EndDate,
		body.RentPerUnit,
		body.Expenses,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: bill})
}
