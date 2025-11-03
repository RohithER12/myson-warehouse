package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterStockRoutes(r *gin.Engine) {
	s := r.Group("/stock")
	s.GET("/", handlers.GetProductStockWithRentHandler)
}
