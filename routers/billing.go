package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func BillingRoutes(r *gin.Engine) {
	b := r.Group("/billing")
	{
		b.POST("/generate", handlers.GenerateBilling)
	}
}
