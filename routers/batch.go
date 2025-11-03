package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

// Optional: register routes
func RegisterBatchRoutes(r *gin.Engine) {
	b := r.Group("/batches")
	{
		b.POST("/", handlers.CreateBatchHandler)
		b.POST("/offboard", handlers.OffboardProductHandler)
		b.GET("/", handlers.GetAllBatchesHandler)
		b.GET("/:id", handlers.GetBatchByIDHandler)
		b.GET("/product/:id", handlers.GetBatchesByProductIDHandler)
		
	}
}
