package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func AnalyticsRoutes(r *gin.Engine) {
	a := r.Group("/analytics")
	{
		a.GET("/products/:id", handlers.GetProductAnalytics)
		a.GET("/products", handlers.GetAllProductsAnalytics)
	}
}
