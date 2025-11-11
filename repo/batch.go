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

// 📦 Get all batches
func (r *BatchRepo) GetAllBatchesCoreData(ctx context.Context) ([]models.BatchCoreData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	type rawData struct {
		ID               uint
		WarehouseID      uint
		StoredAt         time.Time
		Status           string
		CreatedAt        time.Time
		UpdatedAt        time.Time
		BatchStock       int
		AvailableStock   int
		OnBoardedAmount  float64
		OffBoardedAmount float64
	}

	var rows []rawData

	// 🧠 Query combines data from batches, batch_product_entries, and billing_items
	err := db.Table(ns.TableName("Batch") + " AS b").
		Select(`
			b.id,
			b.warehouse_id,
			b.stored_at,
			b.status,
			b.created_at,
			b.updated_at,
			COALESCE(SUM(be.quantity), 0) AS batch_stock,
			COALESCE(SUM(be.stock_quantity), 0) AS available_stock,
			COALESCE(SUM(be.billing_price * be.quantity), 0) AS on_boarded_amount,
			COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0) AS off_boarded_amount
		`).
		Joins("LEFT JOIN " + ns.TableName("BatchProductEntry") + " AS be ON be.batch_id = b.id").
		Joins("LEFT JOIN " + ns.TableName("BillingItem") + " AS bi ON bi.batch_id = b.id").
		Group("b.id, b.warehouse_id, b.stored_at, b.status, b.created_at, b.updated_at").
		Order("b.created_at DESC").
		Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch core data: %w", err)
	}

	var results []models.BatchCoreData
	for _, row := range rows {
		results = append(results, models.BatchCoreData{
			ID:               row.ID,
			WarehouseID:      row.WarehouseID,
			StoredAt:         row.StoredAt,
			Status:           row.Status,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
			BatchStock:       row.BatchStock,
			AvailableStock:   row.AvailableStock,
			OnBoardedAmount:  row.OnBoardedAmount,
			OffBoardedAmount: row.OffBoardedAmount,
		})
	}

	log.Printf("📦 Retrieved %d batch core data records", len(results))
	return results, nil
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
func (r *BatchRepo) GetBatchCoreDataByID(ctx context.Context, id uint) (*models.BatchCoreDataWithProducts, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	// Step 1️⃣: Temporary struct for scanning batch core data only
	type batchCoreRow struct {
		ID               uint
		WarehouseID      uint
		StoredAt         time.Time
		Status           string
		CreatedAt        time.Time
		UpdatedAt        time.Time
		BatchStock       int
		AvailableStock   int
		OffBoardedAmount float64
		OnBoardedAmount  float64
	}

	var batchRow batchCoreRow

	// ✅ Step 2: Fetch batch core data (excluding Product)
	err := db.Table(ns.TableName("Batch") + " AS b").
		Select(`
			b.id,
			b.warehouse_id,
			b.stored_at,
			b.status,
			b.created_at,
			b.updated_at,
			COALESCE(SUM(be.quantity), 0) AS batch_stock,
			COALESCE(SUM(be.stock_quantity), 0) AS available_stock,
			COALESCE(SUM(be.billing_price * be.quantity), 0) AS on_boarded_amount,
			COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0) AS off_boarded_amount
		`).
		Joins("LEFT JOIN " + ns.TableName("BatchProductEntry") + " AS be ON be.batch_id = b.id").
		Joins("LEFT JOIN " + ns.TableName("BillingItem") + " AS bi ON bi.batch_id = b.id").
		Where("b.id = ?", id).
		Group("b.id, b.warehouse_id, b.stored_at, b.status, b.created_at, b.updated_at").
		Scan(&batchRow).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch core data: %w", err)
	}
	if batchRow.ID == 0 {
		return nil, fmt.Errorf("batch not found with id %d", id)
	}

	// ✅ Step 3: Fetch batch products separately
	type productRow struct {
		ProductID      uint
		Name           string
		SupplierID     uint
		Category       string
		StorageArea    float64
		ProductCreated time.Time
		ProductUpdated time.Time
		BillingPrice   float64
		Quantity       int
		StockQuantity  int
		CreatedAt      time.Time
		LastOffboarded *time.Time
		LastUpdated    *time.Time
	}

	var products []productRow

	err = db.Table(ns.TableName("BatchProductEntry") + " AS be").
		Select(`
			be.product_id,
			p.name,
			p.supplier_id,
			p.category,
			p.storage_area,
			p.created_at AS product_created,
			p.updated_at AS product_updated,
			be.billing_price,
			be.quantity,
			be.stock_quantity,
			be.created_at,
			be.last_offboarded,
			be.last_updated
		`).
		Joins("JOIN " + ns.TableName("Product") + " AS p ON be.product_id = p.id").
		Where("be.batch_id = ?", id).
		Order("p.name ASC").
		Scan(&products).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch batch products: %w", err)
	}

	// ✅ Step 4: Map products to model struct
	var batchProducts []models.BatchProductCoreData
	for _, pr := range products {
		batchProducts = append(batchProducts, models.BatchProductCoreData{
			ProductID: pr.ProductID,
			Product: models.ProductCore{
				ID:          pr.ProductID,
				Name:        pr.Name,
				SupplierID:  pr.SupplierID,
				Category:    pr.Category,
				StorageArea: pr.StorageArea,
				CreatedAt:   pr.ProductCreated,
				UpdatedAt:   pr.ProductUpdated,
			},
			BillingPrice:   pr.BillingPrice,
			Quantity:       pr.Quantity,
			StockQuantity:  pr.StockQuantity,
			CreatedAt:      pr.CreatedAt,
			LastOffboarded: pr.LastOffboarded,
			LastUpdated:    pr.LastUpdated,
		})
	}

	// ✅ Step 5: Build final struct safely
	batchCore := models.BatchCoreDataWithProducts{
		ID:               batchRow.ID,
		WarehouseID:      batchRow.WarehouseID,
		StoredAt:         batchRow.StoredAt,
		Status:           batchRow.Status,
		CreatedAt:        batchRow.CreatedAt,
		UpdatedAt:        batchRow.UpdatedAt,
		BatchStock:       batchRow.BatchStock,
		AvailableStock:   batchRow.AvailableStock,
		OffBoardedAmount: batchRow.OffBoardedAmount,
		OnBoardedAmount:  batchRow.OnBoardedAmount,
		Product:          batchProducts,
	}

	log.Printf("📦 Batch %d fetched successfully with %d products", id, len(batchProducts))
	return &batchCore, nil
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
