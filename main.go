package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"warehouse/config"
	dbconn "warehouse/config/dbConn"
	"warehouse/helper"
	routes "warehouse/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ✅ Connect DB
	dbconn.ConnectDB()

	// create Admin
	helper.EnsureAdmin()

	// ✅ Create Gin router
	router := gin.Default()

	// ✅ Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // allow all origins (for dev)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ✅ Setup routes
	routes.SetupRoutes(router)

	// ✅ Start health check goroutine
	go helper.StartHealthPing(config.Cfg.BaseUrl, 30*time.Second)

	// ✅ Create HTTP server
	srv := &http.Server{
		Addr:    ":" + config.Cfg.Port,
		Handler: router,
	}

	// ✅ Run server in a goroutine
	go func() {
		log.Printf("🚀 Server running on port %s\n", config.Cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v\n", err)
		}
	}()

	// ✅ Wait for interrupt signal (Ctrl+C or SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutdown signal received... Cleaning up.")

	// ✅ Gracefully shut down HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println("✅ HTTP server stopped gracefully.")

	// ✅ Close DB connection gracefully
	if dbconn.DB != nil {
		sqlDB, err := dbconn.DB.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("⚠️ Error closing DB: %v", err)
			} else {
				log.Println("✅ Database connection closed.")
			}
		}
	}

	log.Println("👋 Graceful shutdown complete. Exiting.")
}
