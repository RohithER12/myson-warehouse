package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterStockRoutes(r *gin.RouterGroup) {
	s := r.Group("/stock")
	s.GET("/", handlers.GetProductStockWithRentHandler)
	s.GET("/products", handlers.GetAllProductStockHandler)
}
