package repo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"gorm.io/gorm"
)

type BillingRepo struct {
}

// NewBillingRepo initializes the billing repo
func NewBillingRepo() *BillingRepo {
	return &BillingRepo{}
}

// ===============================
// 💳 Create Billing (With BatchID)
// ===============================
func (r *BillingRepo) CreateBillingWithBatchId(ctx context.Context, billingInput models.BillingInput) (*models.Billing, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var billing models.Billing
	err := db.Transaction(func(tx *gorm.DB) error {
		var (
			totalRent, totalStorage, totalBuying, totalSelling, otherExpenses, margin, avgExpense float64
			billingItems                                                                          []models.BillingItem
		)

		// 🧮 Calculate average expense
		if len(billingInput.Expenses) > 0 {
			var expSum float64
			for _, exp := range billingInput.Expenses {
				expSum += exp.Amount
				otherExpenses += exp.Amount
			}
			avgExpense = expSum / float64(len(billingInput.Expenses))
		}

		for _, item := range billingInput.Items {
			var entry models.BatchProductEntry
			if err := tx.Table(ns.TableName("BatchProductEntry")).
				Where("batch_id = ? AND product_id = ?", item.BatchID, item.ProductID).
				First(&entry).Error; err != nil {
				return fmt.Errorf("invalid batch or product reference (batch_id=%v, product_id=%v): %w", item.BatchID, item.ProductID, err)
			}

			if entry.StockQuantity < item.OffboardQty {
				return fmt.Errorf("insufficient stock for product %v in batch %v", item.ProductID, item.BatchID)
			}

			var product models.Product
			if err := tx.Table(ns.TableName("Product")).First(&product, entry.ProductID).Error; err != nil {
				return fmt.Errorf("product not found (ID=%d): %w", entry.ProductID, err)
			}

			var batch models.Batch
			if err := tx.Table(ns.TableName("Batch")).
				Preload("Warehouse.RentConfig").
				First(&batch, item.BatchID).Error; err != nil {
				return fmt.Errorf("batch not found (ID=%v): %w", item.BatchID, err)
			}

			// Rent details
			rate := batch.Warehouse.RentConfig.RatePerSqft
			cycle := strings.ToLower(batch.Warehouse.RentConfig.BillingCycle)

			// Duration calculation
			durationDays := time.Since(batch.CreatedAt).Hours() / 24
			if durationDays < 1 {
				durationDays = 1
			} else if durationDays > 365 {
				durationDays = 365
			}

			var rentMultiplier float64
			switch cycle {
			case "daily":
				rentMultiplier = durationDays
			case "weekly":
				rentMultiplier = durationDays / 7
			case "monthly":
				rentMultiplier = durationDays / 30
			default:
				rentMultiplier = durationDays / 30
			}

			// Cost computations
			areaUsed := product.StorageArea * float64(item.OffboardQty)
			storageCost := rate * areaUsed * rentMultiplier
			totalBuy := float64(item.OffboardQty) * entry.BillingPrice
			totalSell := float64(item.OffboardQty) * item.SellingPrice

			totalStorage += areaUsed
			totalBuying += totalBuy
			totalSelling += totalSell
			totalRent += storageCost

			// ✅ Update stock
			entry.StockQuantity -= item.OffboardQty
			now := time.Now()
			entry.LastOffboarded = &now
			if err := tx.Save(&entry).Error; err != nil {
				return fmt.Errorf("failed to update stock: %w", err)
			}

			// ✅ Update warehouse area
			var warehouse models.Warehouse
			if err := tx.Table(ns.TableName("Warehouse")).First(&warehouse, batch.WarehouseID).Error; err == nil {
				warehouse.AvailableArea += areaUsed
				if warehouse.AvailableArea > warehouse.TotalArea {
					warehouse.AvailableArea = warehouse.TotalArea
				}
				tx.Save(&warehouse)
			}

			// ✅ Record profit
			profit := (item.SellingPrice - entry.BillingPrice) * float64(item.OffboardQty)
			netProfit := profit - storageCost - avgExpense

			if err := tx.Table(ns.TableName("Profit")).Create(&models.Profit{
				BatchID:   entry.BatchID,
				ProductID: entry.ProductID,
				Profit:    profit,
				NetProfit: netProfit,
			}).Error; err != nil {
				return fmt.Errorf("failed to record profit: %w", err)
			}

			// ✅ Add billing item
			billingItems = append(billingItems, models.BillingItem{
				ProductID:    entry.ProductID,
				BatchID:      entry.BatchID,
				OffboardQty:  item.OffboardQty,
				DurationDays: durationDays,
				StorageCost:  storageCost,
				BuyingPrice:  entry.BillingPrice,
				SellingPrice: item.SellingPrice,
				TotalSelling: totalSell,
				BatchStatus:  "offboarded",
			})

			// ✅ Mark batch inactive if all products sold
			var remaining int64
			tx.Table(ns.TableName("BatchProductEntry")).
				Where("batch_id = ? AND stock_quantity > 0", entry.BatchID).
				Count(&remaining)
			if remaining == 0 {
				batch.Status = "inactive"
				tx.Save(&batch)
			}
		}

		// Final margin
		margin = totalSelling - (totalBuying + totalRent + otherExpenses)

		billing = models.Billing{
			Items:         billingItems,
			TotalRent:     totalRent,
			TotalStorage:  totalStorage,
			TotalBuying:   totalBuying,
			TotalSelling:  totalSelling,
			OtherExpenses: otherExpenses,
			Margin:        margin,
		}

		if err := tx.Table(ns.TableName("Billing")).Create(&billing).Error; err != nil {
			return fmt.Errorf("failed to create billing: %w", err)
		}
		
		// -------------------------------------------------------------
		// ⭐ INSERT OFFBOARD EXPENSE ROWS
		// -------------------------------------------------------------
		for _, exp := range billingInput.Expenses {
			offExp := models.OffBoardExpense{
				BillingID: billing.ID,
				Type:      exp.Type,
				Amount:    exp.Amount,
				Notes:     exp.Notes,
			}

			if err := tx.Table(ns.TableName("OffBoardExpense")).Create(&offExp).Error; err != nil {
				return fmt.Errorf("failed to record offboard expense: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Billing creation (BatchID mode) failed: %v", err)
		return nil, err
	}

	log.Printf("✅ Billing created successfully (ID=%d)", billing.ID)
	return &billing, nil
}

// ===============================
// 💳 Create Billing (FIFO Mode)
// ===============================
func (r *BillingRepo) CreateBillingWithOutBatchId(ctx context.Context, billingInput models.BillingInput) (*models.Billing, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var billing models.Billing
	err := db.Transaction(func(tx *gorm.DB) error {
		var (
			totalRent, totalStorage, totalBuying, totalSelling, otherExpenses, margin, avgExpense float64
			billingItems                                                                          []models.BillingItem
		)

		// 🧮 Average expense calculation
		if len(billingInput.Expenses) > 0 {
			var expSum float64
			for _, exp := range billingInput.Expenses {
				expSum += exp.Amount
				otherExpenses += exp.Amount
			}
			avgExpense = expSum / float64(len(billingInput.Expenses))
		}

		for _, item := range billingInput.Items {
			remainingQty := item.OffboardQty
			var batchEntries []models.BatchProductEntry

			// FIFO: fetch oldest first
			if err := tx.Table(ns.TableName("BatchProductEntry")).
				Joins("JOIN "+ns.TableName("Batch")+" AS b ON b.id = "+ns.TableName("BatchProductEntry")+".batch_id").
				Where(ns.TableName("BatchProductEntry")+".product_id = ? AND "+ns.TableName("BatchProductEntry")+".stock_quantity > 0", item.ProductID).
				Order("b.created_at ASC").
				Find(&batchEntries).Error; err != nil {
				return fmt.Errorf("no batches available for product %v: %w", item.ProductID, err)
			}

			if len(batchEntries) == 0 {
				return fmt.Errorf("no available stock for product %v", item.ProductID)
			}

			for _, entry := range batchEntries {
				if remainingQty <= 0 {
					break
				}

				qtyToOffboard := entry.StockQuantity
				if qtyToOffboard > remainingQty {
					qtyToOffboard = remainingQty
				}

				var product models.Product
				if err := tx.Table(ns.TableName("Product")).First(&product, entry.ProductID).Error; err != nil {
					return fmt.Errorf("product not found (ID=%d): %w", entry.ProductID, err)
				}

				var batch models.Batch
				if err := tx.Table(ns.TableName("Batch")).
					Preload("Warehouse.RentConfig").
					First(&batch, entry.BatchID).Error; err != nil {
					return fmt.Errorf("batch not found (ID=%d): %w", entry.BatchID, err)
				}

				rate := batch.Warehouse.RentConfig.RatePerSqft
				cycle := strings.ToLower(batch.Warehouse.RentConfig.BillingCycle)

				durationDays := time.Since(batch.CreatedAt).Hours() / 24
				if durationDays < 1 {
					durationDays = 1
				} else if durationDays > 365 {
					durationDays = 365
				}

				var rentMultiplier float64
				switch cycle {
				case "daily":
					rentMultiplier = durationDays
				case "weekly":
					rentMultiplier = durationDays / 7
				case "monthly":
					rentMultiplier = durationDays / 30
				default:
					rentMultiplier = durationDays / 30
				}

				areaUsed := product.StorageArea * float64(qtyToOffboard)
				storageCost := rate * areaUsed * rentMultiplier
				totalBuy := float64(qtyToOffboard) * entry.BillingPrice
				totalSell := float64(qtyToOffboard) * item.SellingPrice

				totalStorage += areaUsed
				totalBuying += totalBuy
				totalSelling += totalSell
				totalRent += storageCost

				entry.StockQuantity -= qtyToOffboard
				now := time.Now()
				entry.LastOffboarded = &now
				tx.Save(&entry)

				// ✅ Update warehouse
				var warehouse models.Warehouse
				if err := tx.Table(ns.TableName("Warehouse")).First(&warehouse, batch.WarehouseID).Error; err == nil {
					warehouse.AvailableArea += areaUsed
					if warehouse.AvailableArea > warehouse.TotalArea {
						warehouse.AvailableArea = warehouse.TotalArea
					}
					tx.Save(&warehouse)
				}

				// ✅ Record profit
				profit := (item.SellingPrice - entry.BillingPrice) * float64(qtyToOffboard)
				netProfit := profit - storageCost - avgExpense

				if err := tx.Table(ns.TableName("Profit")).Create(&models.Profit{
					BatchID:   entry.BatchID,
					ProductID: entry.ProductID,
					Profit:    profit,
					NetProfit: netProfit,
				}).Error; err != nil {
					return fmt.Errorf("failed to record profit: %w", err)
				}

				billingItems = append(billingItems, models.BillingItem{
					ProductID:    entry.ProductID,
					BatchID:      entry.BatchID,
					OffboardQty:  qtyToOffboard,
					DurationDays: durationDays,
					StorageCost:  storageCost,
					BuyingPrice:  entry.BillingPrice,
					SellingPrice: item.SellingPrice,
					TotalSelling: totalSell,
					BatchStatus:  "offboarded",
				})

				remainingQty -= qtyToOffboard

				// ✅ Mark batch inactive
				var remaining int64
				tx.Table(ns.TableName("BatchProductEntry")).
					Where("batch_id = ? AND stock_quantity > 0", entry.BatchID).
					Count(&remaining)
				if remaining == 0 {
					batch.Status = "inactive"
					tx.Save(&batch)
				}
			}

			if remainingQty > 0 {
				return fmt.Errorf("not enough stock for product %v (needed: %d)", item.ProductID, item.OffboardQty)
			}
		}

		margin = totalSelling - (totalBuying + totalRent + otherExpenses)

		billing = models.Billing{
			Items:         billingItems,
			TotalRent:     totalRent,
			TotalStorage:  totalStorage,
			TotalBuying:   totalBuying,
			TotalSelling:  totalSelling,
			OtherExpenses: otherExpenses,
			Margin:        margin,
		}

		if err := tx.Table(ns.TableName("Billing")).Create(&billing).Error; err != nil {
			return fmt.Errorf("failed to create billing: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Billing creation (FIFO) failed: %v", err)
		return nil, err
	}

	log.Printf("✅ Billing created successfully (FIFO mode, ID=%d)", billing.ID)
	return &billing, nil
}

// ===============================
// 🔍 Get Billing by ID
// ===============================
func (r *BillingRepo) GetByID(ctx context.Context, id uint) (*models.Billing, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var billing models.Billing
	err := db.Table(ns.TableName("Billing")).
		Preload("Items").
		First(&billing, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch billing (ID=%d): %w", id, err)
	}

	return &billing, nil
}
func (r *BillingRepo) GetBillingCoreDataWithProductsByBillID(ctx context.Context, id uint) (models.BillingCoreDataWithProducts, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	// Step 1️⃣: Temporary struct for scalar fields only (no slice)
	type billingRow struct {
		ID            uint
		TotalRent     float64
		TotalStorage  float64
		TotalBuying   float64
		TotalSelling  float64
		OtherExpenses float64
		Margin        float64
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}

	var row billingRow

	// ✅ Step 2: Fetch billing summary
	err := db.Table(ns.TableName("Billing")+" AS b").
		Select(`
			b.id,
			COALESCE(b.total_rent, 0) AS total_rent,
			COALESCE(b.total_storage, 0) AS total_storage,
			COALESCE(b.total_buying, 0) AS total_buying,
			COALESCE(b.total_selling, 0) AS total_selling,
			COALESCE(b.other_expenses, 0) AS other_expenses,
			COALESCE(b.margin, 0) AS margin,
			b.created_at,
			b.updated_at
		`).
		Where("b.id = ?", id).
		Scan(&row).Error

	if err != nil {
		return models.BillingCoreDataWithProducts{}, fmt.Errorf("failed to fetch billing data: %w", err)
	}
	if row.ID == 0 {
		return models.BillingCoreDataWithProducts{}, fmt.Errorf("billing not found with id %d", id)
	}

	// ✅ Step 3: Struct for joined billing items + product info
	type itemRow struct {
		ID               uint
		ProductID        uint
		ProductName      string
		SupplierID       uint
		Category         string
		StorageArea      float64
		ProductCreatedAt time.Time
		ProductUpdatedAt time.Time
		BatchID          uint
		OffboardQty      int
		DurationDays     float64
		StorageCost      float64
		BuyingPrice      float64
		SellingPrice     float64
		TotalSelling     float64
		BatchStatus      string
		CreatedAt        time.Time
		UpdatedAt        time.Time
	}

	var itemRows []itemRow

	// ✅ Step 4: Fetch billing items with joined product data
	err = db.Table(ns.TableName("BillingItem")+" AS bi").
		Select(`
			bi.id,
			bi.product_id,
			p.name AS product_name,
			p.supplier_id,
			p.category,
			p.storage_area,
			p.created_at AS product_created_at,
			p.updated_at AS product_updated_at,
			bi.batch_id,
			bi.offboard_qty,
			bi.duration_days,
			bi.storage_cost,
			bi.buying_price,
			bi.selling_price,
			bi.total_selling,
			bi.batch_status,
			bi.created_at,
			bi.updated_at
		`).
		Joins("JOIN "+ns.TableName("Product")+" AS p ON bi.product_id = p.id").
		Where("bi.billing_id = ?", id).
		Order("bi.created_at ASC").
		Scan(&itemRows).Error

	if err != nil {
		return models.BillingCoreDataWithProducts{}, fmt.Errorf("failed to fetch billing items: %w", err)
	}

	// ✅ Step 5: Build item list with embedded ProductCore
	var items []models.BillingItemCoreData
	for _, ir := range itemRows {
		items = append(items, models.BillingItemCoreData{
			ID: ir.ID,
			Product: models.ProductCore{
				ID:          ir.ProductID,
				Name:        ir.ProductName,
				SupplierID:  ir.SupplierID,
				Category:    ir.Category,
				StorageArea: ir.StorageArea,
				CreatedAt:   ir.ProductCreatedAt,
				UpdatedAt:   ir.ProductUpdatedAt,
			},
			BatchID:      ir.BatchID,
			OffboardQty:  ir.OffboardQty,
			DurationDays: ir.DurationDays,
			StorageCost:  ir.StorageCost,
			BuyingPrice:  ir.BuyingPrice,
			SellingPrice: ir.SellingPrice,
			TotalSelling: ir.TotalSelling,
			BatchStatus:  ir.BatchStatus,
			CreatedAt:    ir.CreatedAt,
			UpdatedAt:    ir.UpdatedAt,
		})
	}

	// ✅ Step 6: Construct the final output struct
	result := models.BillingCoreDataWithProducts{
		ID:            row.ID,
		TotalRent:     row.TotalRent,
		TotalStorage:  row.TotalStorage,
		TotalBuying:   row.TotalBuying,
		TotalSelling:  row.TotalSelling,
		OtherExpenses: row.OtherExpenses,
		Margin:        row.Margin,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		Products:      items,
	}

	log.Printf("🧾 Billing %d fetched successfully with %d products", id, len(items))
	return result, nil
}

// ===============================
// 📋 Get All Billings
// ===============================
func (r *BillingRepo) GetAll(ctx context.Context, warehouseId uint) ([]models.Billing, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var billings []models.Billing

	err := db.Table(ns.TableName("Billing")+" AS bl").
		Joins("LEFT JOIN "+ns.TableName("BillingItem")+" AS bi ON bi.billing_id = bl.id").
		Joins("LEFT JOIN "+ns.TableName("Batch")+" AS b ON b.id = bi.batch_id").
		Where("b.warehouse_id = ?", warehouseId). // ✅ Apply warehouse filter
		Preload("Items").
		Preload("Items.Batch").
		Find(&billings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch billings for warehouse %d: %w", warehouseId, err)
	}

	log.Printf("📑 Retrieved %d billing records for WarehouseID=%d", len(billings), warehouseId)
	return billings, nil
}

func (r *BillingRepo) GetAllBillingCoreData(ctx context.Context, warehouseId uint) ([]models.BillingCoreData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var results []models.BillingCoreData

	err := db.Table(ns.TableName("Billing")+" AS bl").
		Joins("LEFT JOIN "+ns.TableName("BillingItem")+" AS bi ON bi.billing_id = bl.id").
		Joins("LEFT JOIN "+ns.TableName("Batch")+" AS b ON b.id = bi.batch_id").
		Where("b.warehouse_id = ?", warehouseId). // ✅ Warehouse filter
		Select(`
			DISTINCT bl.id,
			COALESCE(bl.total_rent, 0) AS total_rent,
			COALESCE(bl.total_storage, 0) AS total_storage,
			COALESCE(bl.total_buying, 0) AS total_buying,
			COALESCE(bl.total_selling, 0) AS total_selling,
			COALESCE(bl.other_expenses, 0) AS other_expenses,
			COALESCE(bl.margin, 0) AS margin,
			bl.created_at,
			bl.updated_at
		`).
		Order("bl.created_at DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch billing core data for warehouse %d: %w", warehouseId, err)
	}

	log.Printf("🧾 Retrieved %d billing core records for WarehouseID=%d", len(results), warehouseId)
	return results, nil
}

func (r *BillingRepo) GetAllProductsForBilling(ctx context.Context, warehouseId uint) ([]models.ProductStockData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	type rawData struct {
		BatchID         uint
		ProductID       uint
		ProductName     string
		Category        string
		StorageArea     float64
		SupplierID      uint
		SupplierName    string
		SupplierDesc    string
		SupplierCreated time.Time
		SupplierUpdated time.Time
		RentPerSqft     float64
		StockQuantity   int
		BatchCreated    time.Time
		BuyingPrice     float64
		WarehouseID     uint
	}

	var rows []rawData

	// ✅ Apply warehouse filter inside WHERE clause
	err := db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Select(`
			be.batch_id,
			p.id AS product_id,
			p.name AS product_name,
			p.category,
			p.storage_area,
			s.id AS supplier_id,
			s.name AS supplier_name,
			s.description AS supplier_desc,
			s.created_at AS supplier_created,
			s.updated_at AS supplier_updated,
			rr.rate_per_sqft AS rent_per_sqft,
			be.stock_quantity,
			be.billing_price AS buying_price,
			b.created_at AS batch_created,
			b.warehouse_id
		`).
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Joins("JOIN "+ns.TableName("Product")+" AS p ON be.product_id = p.id").
		Joins("JOIN "+ns.TableName("Supplier")+" AS s ON p.supplier_id = s.id").
		Joins("JOIN "+ns.TableName("Warehouse")+" AS w ON b.warehouse_id = w.id").
		Joins("JOIN "+ns.TableName("RentRate")+" AS rr ON w.rent_config_id = rr.id").
		Where(`
			be.stock_quantity > 0 
			AND b.status = 'active'
			AND b.warehouse_id = ?
		`, warehouseId). // ✅ Filtering by warehouse
		Order("b.created_at ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch products for billing: %w", err)
	}

	if len(rows) == 0 {
		log.Printf("⚠️ No active stock available for billing in warehouse %d", warehouseId)
		return nil, nil
	}

	now := time.Now()
	results := make([]models.ProductStockData, 0, len(rows))

	for _, row := range rows {
		// Days in warehouse
		duration := int(now.Sub(row.BatchCreated).Hours() / 24)
		if duration < 1 {
			duration = 1
		}

		// Rent per unit
		rentPerProduct := row.StorageArea * row.RentPerSqft

		productData := models.ProductData{
			ID:         row.ProductID,
			Name:       row.ProductName,
			SupplierID: row.SupplierID,
			Supplier: models.Supplier{
				ID:          row.SupplierID,
				Name:        row.SupplierName,
				Description: row.SupplierDesc,
				CreatedAt:   row.SupplierCreated,
				UpdatedAt:   row.SupplierUpdated,
			},
			Category:    row.Category,
			StorageArea: row.StorageArea,
			CreatedAt:   row.BatchCreated,
			UpdatedAt:   row.BatchCreated,
		}

		expenseData := models.ExpenseData{
			RentPerProduct: rentPerProduct,
			StockQuatity:   row.StockQuantity,
			DurationInDays: duration,
		}

		results = append(results, models.ProductStockData{
			BatchID:     row.BatchID,
			BuyingPrice: row.BuyingPrice,
			ProductData: productData,
			ExpenseData: expenseData,
		})
	}

	log.Printf("📦 Retrieved %d active products for billing in warehouse %d", len(results), warehouseId)
	return results, nil
}
