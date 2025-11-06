package handlers

import (
	"context"
	"net/http"
	"strconv"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var supplierRepo = repo.NewSupplierRepo()

func Createsupplier(c *gin.Context) {
	var p models.Supplier
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	id, err := supplierRepo.Create(context.Background(), &p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: id})
}

func Getsupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	p, err := supplierRepo.GetByID(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: "supplier not found"})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: p})
}

func GetAllsuppliers(c *gin.Context) {
	suppliers, err := supplierRepo.GetAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: suppliers})
}

func Updatesupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	var update models.Supplier
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
update.ID =uint(id)
	err = supplierRepo.Update(context.Background(), update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "supplier updated"})
}

func Deletesupplier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	err = supplierRepo.Delete(context.Background(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.APIResponse{Success: true, Message: "supplier deleted"})
}
