package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func PingRoutes(r *gin.Engine) {
	r.GET("/ping", handlers.PingHandler)
}
