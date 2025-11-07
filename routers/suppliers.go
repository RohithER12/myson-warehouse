package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func SupplierRoutes(r *gin.RouterGroup) {
	p := r.Group("/suppliers")
	{
		p.POST("/", handlers.Createsupplier)
		p.GET("/", handlers.GetAllsuppliers)
		p.GET("/:id", handlers.Getsupplier)
		p.PUT("/:id", handlers.Updatesupplier)
		p.DELETE("/:id", handlers.Deletesupplier)
	}
}
