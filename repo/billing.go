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
			billingItems                                                                         []models.BillingItem
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
				return fmt.Errorf("insufficient stock for product %d in batch %d", item.ProductID, item.BatchID)
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
			billingItems                                                                         []models.BillingItem
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
				return fmt.Errorf("no batches available for product %d: %w", item.ProductID, err)
			}

			if len(batchEntries) == 0 {
				return fmt.Errorf("no available stock for product %d", item.ProductID)
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
				return fmt.Errorf("not enough stock for product %d (needed: %d)", item.ProductID, item.OffboardQty)
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

// ===============================
// 📋 Get All Billings
// ===============================
func (r *BillingRepo) GetAll(ctx context.Context) ([]models.Billing, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var billings []models.Billing
	err := db.Table(ns.TableName("Billing")).
		Preload("Items").
		Find(&billings).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch billings: %w", err)
	}

	log.Printf("📑 Retrieved %d billing records", len(billings))
	return billings, nil
}