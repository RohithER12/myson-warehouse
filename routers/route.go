package routes

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Register all resource routes
	WarehouseRoutes(r)
	ProductRoutes(r)
	BillingRoutes(r)
	AnalyticsRoutes(r)
	RegisterBatchRoutes(r)
	SupplierRoutes(r)
	RegisterStockRoutes(r)
}
