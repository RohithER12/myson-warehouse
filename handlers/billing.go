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
		Items       []repo.BillingItemInput `json:"items"`
		EndDate     time.Time               `json:"end_date"`
		RentPerUnit float64                 `json:"rent_per_unit"`
		Expenses    []models.Expense        `json:"expenses"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	bill, err := billingRepo.GenerateBilling(
		context.Background(),
		body.Items,
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
