package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	u := r.Group("/user")
	{
		u.POST("/login", handlers.Login)
		u.POST("/register", handlers.Register)
	}
}
