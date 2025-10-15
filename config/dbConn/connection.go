package dbconn

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"warehouse/config"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var (
	clientInstance *mongo.Client
	once           sync.Once
)

// InitMongoClient initializes and returns a singleton MongoDB client.
func InitMongoClient() *mongo.Client {
	config.LoadConfig()
	once.Do(func() {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(config.Cfg.DBConnectionString).SetServerAPIOptions(serverAPI)

		c, err := mongo.Connect(opts)
		if err != nil {
			log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
		}

		// Ping the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.Ping(ctx, readpref.Primary()); err != nil {
			log.Fatalf("❌ Failed to ping MongoDB: %v", err)
		}

		fmt.Println("✅ Connected to MongoDB Atlas")
		clientInstance = c
	})
	return clientInstance
}

// GetCollection returns a collection from the connected MongoDB client.
func GetCollection(dbName, collectionName string) *mongo.Collection {
	client := InitMongoClient()
	if client == nil {
		log.Fatal("❌ MongoDB client is not initialized")
	}
	return client.Database(dbName).Collection(collectionName)
}

// GetClient returns the MongoDB client instance
func GetClient() *mongo.Client {
	if clientInstance == nil {
		// Ensure client is initialized
		clientInstance = InitMongoClient()
	}
	return clientInstance
}
