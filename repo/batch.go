package repo

import (
	"context"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type BatchRepo struct{}

func NewBatchRepo() *BatchRepo {
	return &BatchRepo{}
}

// ➕ Add new batch
func (r *BatchRepo) AddBatch(ctx context.Context, batch *models.Batch) (uint, error) {
	batch.Status = "active"
	now := time.Now()
	// Set stock quantities
	for i := range batch.Products {
		batch.Products[i].StockQuantity = batch.Products[i].Quantity
		batch.Products[i].LastUpdated = &now

	}
	if dbconn.DB == nil {
		dbconn.ConnectDB()
	}
	if err := dbconn.DB.WithContext(ctx).Create(batch).Error; err != nil {
		return 0, err
	}
	return batch.ID, nil
}

// 📦 Get all batches
func (r *BatchRepo) GetAllBatches(ctx context.Context) ([]models.Batch, error) {
	var batches []models.Batch

	err := dbconn.DB.WithContext(ctx).
		Preload("Warehouse.RentConfig").
		Preload("Products.Product.Supplier").
		Find(&batches).Error

	return batches, err
}

// 🔍 Get batch by ID
func (r *BatchRepo) GetBatchByID(ctx context.Context, id uint) (*models.Batch, error) {
	var batch models.Batch
	err := dbconn.DB.WithContext(ctx).Preload("Warehouse.RentConfig").
		Preload("Products.Product.Supplier").
		Find(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// 🔍 Get batches by Product ID
func (r *BatchRepo) GetBatchesByProductID(ctx context.Context, productID string) ([]models.Batch, error) {
	var batches []models.Batch

	err := dbconn.DB.WithContext(ctx).
		Joins("JOIN batch_product_entries ON batches.id = batch_product_entries.batch_id").
		Where("batch_product_entries.product_id = ?", productID).
		Preload("Products.Product.Supplier").
		Preload("Warehouse.RentConfig").
		Find(&batches).Error

	return batches, err
}
