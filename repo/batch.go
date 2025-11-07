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
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	batch.Status = "active"
	now := time.Now()
	var returnID uint

	err := db.Transaction(func(tx *gorm.DB) error {
		// Step 1️⃣: Load warehouse
		var warehouse models.Warehouse
		if err := tx.Table(ns.TableName("Warehouse")).
			Preload("RentConfig").
			First(&warehouse, batch.WarehouseID).Error; err != nil {
			return fmt.Errorf("warehouse not found with ID %d", batch.WarehouseID)
		}

		// Step 2️⃣: Calculate total used space
		var totalUsedArea float64
		for i := range batch.Products {
			productEntry := &batch.Products[i]

			// Fetch product to get its storage area
			var product models.Product
			if err := tx.Table(ns.TableName("Product")).
				First(&product, productEntry.ProductID).Error; err != nil {
				return fmt.Errorf("product not found for ID %d", productEntry.ProductID)
			}

			// Initialize stock info
			productEntry.StockQuantity = productEntry.Quantity
			productEntry.LastUpdated = &now

			// Calculate space usage
			usedArea := product.StorageArea * float64(productEntry.Quantity)
			totalUsedArea += usedArea

			log.Printf("📦 Product %d → Qty %d × %.2f sqft = %.2f sqft used",
				productEntry.ProductID, productEntry.Quantity, product.StorageArea, usedArea)
		}

		// Step 3️⃣: Validate warehouse space
		if warehouse.AvailableArea < totalUsedArea {
			return fmt.Errorf("❌ insufficient warehouse space (available: %.2f, required: %.2f sqft)",
				warehouse.AvailableArea, totalUsedArea)
		}

		// Step 4️⃣: Deduct used area and update warehouse
		warehouse.AvailableArea -= totalUsedArea
		if err := tx.Table(ns.TableName("Warehouse")).
			Save(&warehouse).Error; err != nil {
			return fmt.Errorf("failed to update warehouse space: %w", err)
		}

		// Step 5️⃣: Create batch
		batch.StoredAt = now
		if err := tx.Table(ns.TableName("Batch")).
			Create(batch).Error; err != nil {
			return fmt.Errorf("failed to create batch: %w", err)
		}

		returnID = batch.ID
		log.Printf("✅ Batch created (ID=%d) | Used %.2f sqft | Warehouse %d", batch.ID, totalUsedArea, batch.WarehouseID)
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
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var batches []models.Batch
	err := db.Table(ns.TableName("Batch")).
		Preload("Warehouse.RentConfig").
		Preload("Products.Product.Supplier").
		Find(&batches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batches: %w", err)
	}

	log.Printf("📦 Retrieved %d batches", len(batches))
	return batches, nil
}

// 🔍 Get batch by ID
func (r *BatchRepo) GetBatchByID(ctx context.Context, id uint) (*models.Batch, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var batch models.Batch
	err := db.Table(ns.TableName("Batch")).
		Preload("Warehouse.RentConfig").
		Preload("Products.Product.Supplier").
		First(&batch, id).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch ID %d: %w", id, err)
	}

	log.Printf("🔍 Retrieved batch ID=%d", id)
	return &batch, nil
}

// 🔍 Get batches by Product ID
func (r *BatchRepo) GetBatchesByProductID(ctx context.Context, productID string) ([]models.Batch, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var batches []models.Batch
	err := db.Table(ns.TableName("Batch")).
		Joins("JOIN "+ns.TableName("BatchProductEntry")+" AS bpe ON bpe.batch_id = "+ns.TableName("Batch")+".id").
		Where("bpe.product_id = ?", productID).
		Preload("Products.Product.Supplier").
		Preload("Warehouse.RentConfig").
		Find(&batches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batches for product %s: %w", productID, err)
	}

	log.Printf("🔍 Retrieved %d batches for ProductID=%s", len(batches), productID)
	return batches, nil
}