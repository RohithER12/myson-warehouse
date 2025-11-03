package main

import (
	"context"
	"log"
	"warehouse/config"
	dbconn "warehouse/config/dbConn"
	routes "warehouse/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// log.Println("DBConnectionString :", config.Cfg.DBConnectionString)
	dbconn.InitMongoClient()

	dbconn.Migrate(config.Cfg.DBName)

	log.Println("🚀 Migration done, DB ready.")
	defer dbconn.GetClient().Disconnect(context.TODO())

	r := gin.Default()
	setupRoutes(r)
	r.Run(":" + config.Cfg.Port)
}

func setupRoutes(r *gin.Engine) {
	// Register all resource routes
	routes.WarehouseRoutes(r)
	routes.ProductRoutes(r)
	routes.BillingRoutes(r)
	routes.AnalyticsRoutes(r)
	routes.RegisterBatchRoutes(r)
	routes.SupplierRoutes(r)
	routes.RegisterStockRoutes(r)
}
