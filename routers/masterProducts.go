package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func ProductRoutes(r *gin.RouterGroup) {
	p := r.Group("/products")
	{
		p.POST("/", handlers.CreateProduct)
		p.GET("/", handlers.GetAllProducts)
		p.GET("/categories", handlers.GetAllProductCategories)
		p.GET("/:id", handlers.GetProduct)
		p.PUT("/:id", handlers.UpdateProduct)
		p.DELETE("/:id", handlers.DeleteProduct)
	}
}
