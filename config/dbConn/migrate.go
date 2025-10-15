package dbconn

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Migrate(dbName string) {
	client := GetClient()
	db := client.Database(dbName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collections := []string{
		"warehouses",
		"products",
		"billings",
		"analytics",
	}

	for _, name := range collections {
		createCollectionIfNotExists(ctx, db, name)
	}

	createIndexes(ctx, db)
	log.Println("✅ MongoDB migration completed")
}

func createCollectionIfNotExists(ctx context.Context, db *mongo.Database, name string) {
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": name})
	if err != nil {
		log.Printf("⚠️ Failed to list collections: %v", err)
		return
	}

	if len(collections) == 0 {
		err = db.CreateCollection(ctx, name)
		if err != nil {
			log.Printf("⚠️ Failed to create collection '%s': %v", name, err)
			return
		}
		fmt.Printf("📦 Created collection: %s\n", name)
	}
}

func createIndexes(ctx context.Context, db *mongo.Database) {
	// Index for fast lookup
	productCol := db.Collection("products")
	_, err := productCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "warehouse_id", Value: 1},
		},
	})
	if err != nil {
		log.Printf("⚠️ Failed to create product index: %v", err)
	}

	billingCol := db.Collection("billings")
	_, err = billingCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "product_id", Value: 1}},
	})
	if err != nil {
		log.Printf("⚠️ Failed to create billing index: %v", err)
	}
}
