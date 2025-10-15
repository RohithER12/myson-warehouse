package repo

import (
	"context"
	"fmt"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const BatchCollection = "batches"

type StockRepo struct {
	col *mongo.Collection
}

// NewStockRepo initializes the repository
func NewStockRepo() *StockRepo {
	return &StockRepo{
		col: dbconn.GetCollection("myson_warehouse", productCollection),
	}
}

func (r *StockRepo) AddBatch(ctx context.Context, batch *models.Batch) (string, error) {
	batch.ID = primitive.NewObjectID()
	batch.Status = "active"
	batch.StoredAt = time.Now()

	_, err := r.col.InsertOne(ctx, batch)
	if err != nil {
		return "", err
	}

	log.Printf("📦 New batch created: %s in warehouse %s", batch.ID.Hex(), batch.WarehouseID)
	return batch.ID.Hex(), nil
}

func (r *StockRepo) Offboard(ctx context.Context, productID string, offboardQty int) error {
	objID, _ := primitive.ObjectIDFromHex(productID)
	remaining := offboardQty

	// Fetch active batches that contain the product, sorted by StoredAt (FIFO)
	cursor, err := r.col.Find(ctx, bson.M{
		"products.product_id": objID,
		"status":              "active",
	}, options.Find().SetSort(bson.M{"stored_at": 1}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var batch models.Batch
		if err := cursor.Decode(&batch); err != nil {
			return err
		}

		for i, p := range batch.Products {
			if p.ProductID != objID || p.Quantity == 0 {
				continue
			}

			deduct := p.Quantity
			if p.Quantity > remaining {
				deduct = remaining
			}

			batch.Products[i].Quantity -= deduct
			now := time.Now()
			batch.Products[i].LastOffboarded = &now

			remaining -= deduct
			if remaining <= 0 {
				break
			}
		}

		// Update batch status
		allZero := true
		for _, p := range batch.Products {
			if p.Quantity > 0 {
				allZero = false
				break
			}
		}
		if allZero {
			batch.Status = "fully_offboarded"
		} else {
			batch.Status = "partially_offboarded"
		}

		_, err := r.col.UpdateOne(ctx, bson.M{"_id": batch.ID}, bson.M{"$set": batch})
		if err != nil {
			return err
		}

		if remaining <= 0 {
			break
		}
	}

	if remaining > 0 {
		return fmt.Errorf("not enough quantity in batches for product %s", productID)
	}
	return nil
}

func (r *StockRepo) GetAllBatches(ctx context.Context) ([]models.Batch, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []models.Batch
	for cursor.Next(ctx) {
		var batch models.Batch
		if err := cursor.Decode(&batch); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}

	return batches, nil
}

func (r *StockRepo) GetBatchByID(ctx context.Context, batchID string) (*models.Batch, error) {
	objID, err := primitive.ObjectIDFromHex(batchID)
	if err != nil {
		return nil, fmt.Errorf("invalid batch ID: %v", err)
	}

	var batch models.Batch
	if err := r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&batch); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("batch not found")
		}
		return nil, err
	}

	return &batch, nil
}

func (r *StockRepo) GetBatchesByProductID(ctx context.Context, productID string) ([]models.Batch, error) {
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %v", err)
	}

	cursor, err := r.col.Find(ctx, bson.M{"products.product_id": objID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var batches []models.Batch
	for cursor.Next(ctx) {
		var batch models.Batch
		if err := cursor.Decode(&batch); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}

	return batches, nil
}

func (r *StockRepo) UpdateBatch(ctx context.Context, batch *models.Batch) error {
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": batch.ID},
		bson.M{"$set": batch},
	)
	return err
}
