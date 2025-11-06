package handlers

import (
	"context"
	"net/http"
	"strconv"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var warehouseRepo = repo.NewWarehouseRepo()

func CreateWarehouse(c *gin.Context) {
	var wh models.Warehouse
	if err := c.ShouldBindJSON(&wh); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	id, err := warehouseRepo.Create(context.Background(), &wh)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: id})
}

func GetWarehouse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	wh, err := warehouseRepo.GetByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "Warehouse not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: wh})
}

func GetAllWarehouses(c *gin.Context) {
	whs, err := warehouseRepo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: whs})
}

func UpdateWarehouse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	var update models.Warehouse
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	update.ID = uint(id)
	err = warehouseRepo.Update(context.Background(), &update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Warehouse updated"})
}

func DeleteWarehouse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	err = warehouseRepo.Delete(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "Warehouse deleted"})
}
