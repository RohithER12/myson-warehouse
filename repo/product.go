package repo

import (
	"context"
	"log"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const productCollection = "products"

type ProductRepo struct {
	col *mongo.Collection
}

// NewProductRepo initializes the repository
func NewProductRepo() *ProductRepo {
	return &ProductRepo{
		col: dbconn.GetCollection("myson_warehouse", productCollection),
	}
}

// Create inserts a new product
func (r *ProductRepo) Create(ctx context.Context, product *models.Product) (string, error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = primitive.NewObjectID()

	_, err := r.col.InsertOne(ctx, product)
	if err != nil {
		return "", err
	}

	log.Printf("🛠 New product created: %s", product.ID.Hex())
	return product.ID.Hex(), nil
}

// GetByID fetches a product by ID
func (r *ProductRepo) GetByID(ctx context.Context, id string) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var product models.Product
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetAll fetches all products
func (r *ProductRepo) GetAll(ctx context.Context) ([]models.Product, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []models.Product
	for cursor.Next(ctx) {
		var p models.Product
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// Update modifies a product
func (r *ProductRepo) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update["last_updated"] = time.Now()
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

// Delete removes a product
func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
