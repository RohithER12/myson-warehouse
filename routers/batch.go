package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

// Optional: register routes
func RegisterBatchRoutes(r *gin.RouterGroup) {
	b := r.Group("/batches")
	{
		b.POST("/", handlers.CreateBatchHandler)
		b.GET("/", handlers.GetAllBatchesHandler)
		b.GET("/:id", handlers.GetBatchByIDHandler)
		b.GET("/product/:id", handlers.GetBatchesByProductIDHandler)
		
	}
}
