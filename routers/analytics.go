package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func AnalyticsRoutes(r *gin.RouterGroup) {
	a := r.Group("/analytics")
	{
		a.GET("/:duration", handlers.GetAnalyticsHandler)
	}
}
