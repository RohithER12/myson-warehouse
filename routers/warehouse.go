package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func WarehouseRoutes(r *gin.Engine) {
	w := r.Group("/warehouses")
	{
		w.POST("/", handlers.CreateWarehouse)
		w.GET("/", handlers.GetAllWarehouses)
		w.GET("/:id", handlers.GetWarehouse)
		w.PUT("/:id", handlers.UpdateWarehouse)
		w.DELETE("/:id", handlers.DeleteWarehouse)
	}
}
