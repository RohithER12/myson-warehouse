package routes

import (
	"warehouse/handlers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	u := r.Group("/user")
	{
		u.POST("/login", handlers.Login)
		u.POST("/register", handlers.RegisterUser)
	}
}
func AdminRoutes(r *gin.RouterGroup) {
	u := r.Group("/admin")
	{
		u.POST("/register", handlers.RegisterAdmin)
	}
}
