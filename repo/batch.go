package repo

import (
	"context"
	"fmt"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"gorm.io/gorm"
)

type BatchRepo struct{}

func NewBatchRepo() *BatchRepo {
	return &BatchRepo{}
}

// ➕ Add new batch
func (r *BatchRepo) AddBatch(ctx context.Context, batch *models.Batch) (uint, error) {
	batch.Status = "active"
	now := time.Now()

	returnID := uint(0)

	err := dbconn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1️⃣: Load warehouse
		var warehouse models.Warehouse
		if err := tx.Preload("RentConfig").First(&warehouse, batch.WarehouseID).Error; err != nil {
			return fmt.Errorf("warehouse not found with ID %d", batch.WarehouseID)
		}

		// Step 2️⃣: Calculate total area needed
		var totalUsedArea float64
		for i := range batch.Products {
			productEntry := &batch.Products[i]

			// Fetch product to get storage area
			var product models.Product
			if err := tx.First(&product, productEntry.ProductID).Error; err != nil {
				return fmt.Errorf("product not found for ID %d", productEntry.ProductID)
			}

			// Set initial stock
			productEntry.StockQuantity = productEntry.Quantity
			productEntry.LastUpdated = &now

			// Calculate space usage
			usedArea := product.StorageArea * float64(productEntry.Quantity)
			totalUsedArea += usedArea

			log.Printf("📦 Product ID %d → Qty %d × %.2f sqft = %.2f sqft used",
				productEntry.ProductID, productEntry.Quantity, product.StorageArea, usedArea)
		}

		// Step 3️⃣: Validate available warehouse space
		if warehouse.AvailableArea < totalUsedArea {
			return fmt.Errorf("❌ insufficient warehouse space: available %.2f, required %.2f sqft",
				warehouse.AvailableArea, totalUsedArea)
		}

		// Step 4️⃣: Deduct used area and update warehouse
		warehouse.AvailableArea -= totalUsedArea
		if err := tx.Save(&warehouse).Error; err != nil {
			return fmt.Errorf("failed to update warehouse available area: %v", err)
		}

		// Step 5️⃣: Create the batch record
		batch.StoredAt = time.Now()
		if err := tx.Create(batch).Error; err != nil {
			return fmt.Errorf("failed to create batch: %v", err)
		}

		returnID = batch.ID
		log.Printf("✅ Batch created successfully (ID: %d), used %.2f sqft of warehouse %d", batch.ID, totalUsedArea, batch.WarehouseID)
		return nil
	})

	if err != nil {
		log.Printf("🔥 Failed to add batch: %v", err)
		return 0, err
	}

	return returnID, nil
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
