package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func BillingRoutes(r *gin.Engine) {
	b := r.Group("/billing")
	{
		b.POST("/generate/batch", handlers.CreateBillingWithBatchId)
		b.POST("/generate", handlers.CreateBillingWithOutBatchId)
		b.GET("/", handlers.GetAllBillsHandler)
		b.GET("/:id", handlers.GetBillByIDHandler)
	}
}
