package handlers

import (
	"context"
	"net/http"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var productRepo = repo.NewProductRepo()

func CreateProduct(c *gin.Context) {
	var p models.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	id, err := productRepo.Create(context.Background(), &p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: id})
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")
	p, err := productRepo.GetByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Product not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: p})
}

func GetAllProducts(c *gin.Context) {
	products, err := productRepo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: products})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var update map[string]interface{}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	err := productRepo.Update(context.Background(), id, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Product updated"})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	err := productRepo.Delete(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Product deleted"})
}

func OffboardProduct(c *gin.Context) {
	productID := c.Param("id")

	var body struct {
		Quantity int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	if body.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: "Quantity must be greater than 0"})
		return
	}

	// Offboard quantity from batches
	if err := stockRepo.Offboard(context.Background(), productID, body.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Product offboarded successfully"})
}
