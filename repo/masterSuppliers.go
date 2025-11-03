package repo

import (
	"context"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)



type SupplierRepo struct {
	col *mongo.Collection
}

// NewSupplierRepo initializes the repository
func NewSupplierRepo() *SupplierRepo {
	return &SupplierRepo{
		col: dbconn.GetCollection("myson_warehouse", SupplierCollection),
	}
}

// Create inserts a new supplier
func (r *SupplierRepo) Create(ctx context.Context, supplier *models.Supplier) (string, error) {
	supplier.CreatedAt = time.Now()
	supplier.UpdatedAt = time.Now()
	supplier.ID = primitive.NewObjectID()

	_, err := r.col.InsertOne(ctx, supplier)
	if err != nil {
		return "", err
	}

	log.Printf("🛠 New supplier created: %s", supplier.ID.Hex())
	return supplier.ID.Hex(), nil
}

// GetByID fetches a supplier by ID
func (r *SupplierRepo) GetByID(ctx context.Context, id string) (*models.Supplier, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var supplier models.Supplier
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&supplier)
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetAll fetches all suppliers
func (r *SupplierRepo) GetAll(ctx context.Context) ([]models.Supplier, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var suppliers []models.Supplier
	for cursor.Next(ctx) {
		var p models.Supplier
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, p)
	}
	return suppliers, nil
}

// Update modifies a supplier
func (r *SupplierRepo) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update["updated_at"] = time.Now()
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

// Delete removes a supplier
func (r *SupplierRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
