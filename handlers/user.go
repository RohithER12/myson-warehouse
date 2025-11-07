package handlers

import (
	"net/http"
	"warehouse/helper"
	"warehouse/models"
	"warehouse/repo"

	"github.com/gin-gonic/gin"
)

var userRepo = repo.NewUserRepo()

func Register(c *gin.Context) {
	var in models.RegisterDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if in.ConfirmPassword != in.Password {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "password mismatch"})
		return
	}
	hash, err := helper.HashPassword(in.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "hash error"})
		return
	}
	u := &models.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
		WarehouseID:  in.WarehouseID,
		Role:         models.RoleEmployee,
	}
	if err := userRepo.Create(c, u); err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": u.ID}})
}

func Login(c *gin.Context) {
	var in models.LoginDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	u, err := userRepo.GetByEmail(c, in.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid credentials"})
		return
	}
	if err := helper.CheckPassword(u.PasswordHash, in.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "invalid credentials"})
		return
	}

	token, err := helper.CreateJWT(u.ID, u.WarehouseID, u.Email, string(u.Role), 60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "token creation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id": u.ID, "name": u.Name, "email": u.Email, "role": u.Role, "warehouse_id": u.WarehouseID,
			},
		},
	})
}
