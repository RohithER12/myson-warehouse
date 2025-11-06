package main

import (
	"log"
	"time"
	"warehouse/config"
	dbconn "warehouse/config/dbConn"
	routes "warehouse/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Connect DB
	dbconn.ConnectDB()

	// Server
	route := gin.Default()

	// ✅ Setup CORS
	route.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // ✅ allow all origins
		}, // change to your frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.SetupRoutes(route)

	log.Printf("🚀 Server running on port %s\n", config.Cfg.Port)
	route.Run(":" + config.Cfg.Port)
}
