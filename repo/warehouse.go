package repo

import (
	"context"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const warehouseCollection = "warehouses"

type WarehouseRepo struct {
	col *mongo.Collection
}

// NewWarehouseRepo initializes the repository
func NewWarehouseRepo() *WarehouseRepo {
	return &WarehouseRepo{
		col: dbconn.GetCollection("myson_warehouse", warehouseCollection),
	}
}

// Create inserts a new warehouse
func (r *WarehouseRepo) Create(ctx context.Context, warehouse *models.Warehouse) (string, error) {
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	// Generate ObjectID for new warehouse
	if warehouse.ID.IsZero() {
		warehouse.ID = primitive.NewObjectID()
	}

	_, err := r.col.InsertOne(ctx, warehouse)
	if err != nil {
		return "", err
	}

	log.Printf("🏠 New warehouse created: %s", warehouse.ID.Hex())
	return warehouse.ID.Hex(), nil
}

// GetByID fetches a warehouse by ID
func (r *WarehouseRepo) GetByID(ctx context.Context, id string) (*models.Warehouse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var warehouse models.Warehouse
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&warehouse)
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// GetAll fetches all warehouses
func (r *WarehouseRepo) GetAll(ctx context.Context) ([]models.Warehouse, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var warehouses []models.Warehouse
	for cursor.Next(ctx) {
		var w models.Warehouse
		if err := cursor.Decode(&w); err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

// Update modifies a warehouse record
func (r *WarehouseRepo) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update["updated_at"] = time.Now()

	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

// Delete removes a warehouse
func (r *WarehouseRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
