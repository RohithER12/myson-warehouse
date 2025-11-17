package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func AnalyticsRoutes(r *gin.RouterGroup) {
	a := r.Group("/analytics")
	{
		a.GET("/:duration", handlers.GetAnalyticsHandler)
		a.GET("/fast-moving", handlers.GetFastAndSlowMovingProductAnalytics)
		a.GET("/product/:product_id", handlers.GetProductAnalyticsByIdHandler)
	}
}
