package routes

import (
	"warehouse/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	
	// Public routes
	UserRoutes(r)
	PingRoutes(r)

	// Protected routes
	api := r.Group("/")
	api.Use(middleware.AuthJWT())

	// Employee + Admin routes
	registerEmployeeAndAdminRoutes(api)

	// Admin-only routes
	registerAdminRoutes(api)
}

func registerEmployeeAndAdminRoutes(r *gin.RouterGroup) {
	group := r.Group("/")
	group.Use(middleware.RequireRoles("employee", "admin"))

	WarehouseRoutes(group)
	ProductRoutes(group)
	BillingRoutes(group)
	RegisterBatchRoutes(group)
	SupplierRoutes(group)
	RegisterStockRoutes(group)
}

func registerAdminRoutes(r *gin.RouterGroup) {
	admin := r.Group("/")
	admin.Use(middleware.RequireRoles("admin"))

	AnalyticsRoutes(admin)
}
