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

func (r *BillingRepo) CreateBillingWithBatchId(ctx context.Context, billingInput models.BillingInput) (*models.Billing, error) {
	var billing models.Billing

	err := dbconn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var totalRent, totalStorage, totalBuying, totalSelling, otherExpenses, margin, avgExpense float64
		var billingItems []models.BillingItem

		// 🧮 Calculate average expense for this billing
		if len(billingInput.Expenses) > 0 {
			var expSum float64
			for _, exp := range billingInput.Expenses {
				expSum += exp.Amount
				otherExpenses += exp.Amount
			}
			avgExpense = expSum / float64(len(billingInput.Expenses))
		}

		for _, item := range billingInput.Items {

			// Fetch batch-product entry
			var entry models.BatchProductEntry
			if err := tx.Where("batch_id = ? AND product_id = ?", item.BatchID, item.ProductID).First(&entry).Error; err != nil {
				return fmt.Errorf("invalid batch or product reference for batch_id=%v product_id=%v", item.BatchID, item.ProductID)
			}

			if entry.StockQuantity < item.OffboardQty {
				return fmt.Errorf("insufficient stock for product ID %v in batch %v", item.ProductID, item.BatchID)
			}

			// Fetch product
			var product models.Product
			if err := tx.First(&product, entry.ProductID).Error; err != nil {
				return fmt.Errorf("product not found for ID %d", entry.ProductID)
			}

			// Fetch batch + warehouse + rent config
			var batch models.Batch
			if err := tx.Preload("Warehouse.RentConfig").First(&batch, item.BatchID).Error; err != nil {
				return fmt.Errorf("batch not found for ID %v", item.BatchID)
			}

			rate := batch.Warehouse.RentConfig.RatePerSqft
			cycle := strings.ToLower(batch.Warehouse.RentConfig.BillingCycle)

			// Calculate rent duration
			durationDays := time.Since(batch.CreatedAt).Hours() / 24
			if durationDays < 1 {
				durationDays = 1
			}
			if durationDays > 365 {
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

			// Calculate costs
			areaUsed := product.StorageArea * float64(item.OffboardQty)
			storageCost := rate * areaUsed * rentMultiplier
			totalBuy := float64(item.OffboardQty) * entry.BillingPrice
			totalSell := float64(item.OffboardQty) * item.SellingPrice

			totalStorage += areaUsed
			totalBuying += totalBuy
			totalSelling += totalSell
			totalRent += storageCost

			// ✅ Update stock quantity
			entry.StockQuantity -= item.OffboardQty
			now := time.Now()
			entry.LastOffboarded = &now
			if err := tx.Save(&entry).Error; err != nil {
				return err
			}

			// ✅ Update warehouse space
			var warehouse models.Warehouse
			if err := tx.First(&warehouse, batch.WarehouseID).Error; err == nil {
				warehouse.AvailableArea += areaUsed
				if warehouse.AvailableArea > warehouse.TotalArea {
					warehouse.AvailableArea = warehouse.TotalArea
				}
				if err := tx.Save(&warehouse).Error; err != nil {
					return fmt.Errorf("failed to update warehouse space: %v", err)
				}
			}

			// ✅ Calculate profit
			profit := (item.SellingPrice - entry.BillingPrice) * float64(item.OffboardQty)
			netProfit := profit - storageCost - avgExpense

			if err := tx.Create(&models.Profit{
				BatchID:   entry.BatchID,
				ProductID: entry.ProductID,
				Profit:    profit,
				NetProfit: netProfit,
			}).Error; err != nil {
				return fmt.Errorf("failed to record profit: %v", err)
			}

			// Add billing item
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
			tx.Model(&models.BatchProductEntry{}).
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

		if err := tx.Create(&billing).Error; err != nil {
			return fmt.Errorf("failed to create billing: %v", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Transaction failed: %v", err)
	}
	return &billing, err
}

// FIFO ALG
func (r *BillingRepo) CreateBillingWithOutBatchId(ctx context.Context, billingInput models.BillingInput) (*models.Billing, error) {
	var billing models.Billing

	err := dbconn.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var totalRent, totalStorage, totalBuying, totalSelling, otherExpenses, margin, avgExpense float64
		var billingItems []models.BillingItem

		// 🧮 Compute average expense once
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

			// FIFO: oldest batches first
			if err := tx.
				Joins("JOIN batches ON batches.id = batch_product_entries.batch_id").
				Where("batch_product_entries.product_id = ? AND batch_product_entries.stock_quantity > 0", item.ProductID).
				Order("batches.created_at ASC").
				Find(&batchEntries).Error; err != nil {
				return fmt.Errorf("no available batches found for product ID %v", item.ProductID)
			}

			if len(batchEntries) == 0 {
				return fmt.Errorf("no available stock for product ID %v", item.ProductID)
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
				if err := tx.First(&product, entry.ProductID).Error; err != nil {
					return fmt.Errorf("product not found for ID %d", entry.ProductID)
				}

				var batch models.Batch
				if err := tx.Preload("Warehouse.RentConfig").First(&batch, entry.BatchID).Error; err != nil {
					return fmt.Errorf("batch not found for ID %d", entry.BatchID)
				}

				rate := batch.Warehouse.RentConfig.RatePerSqft
				cycle := strings.ToLower(batch.Warehouse.RentConfig.BillingCycle)

				durationDays := time.Since(batch.CreatedAt).Hours() / 24
				if durationDays < 1 {
					durationDays = 1
				}
				if durationDays > 365 {
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

				// ✅ Update stock
				entry.StockQuantity -= qtyToOffboard
				now := time.Now()
				entry.LastOffboarded = &now
				if err := tx.Save(&entry).Error; err != nil {
					return err
				}

				// ✅ Update warehouse available area
				var warehouse models.Warehouse
				if err := tx.First(&warehouse, batch.WarehouseID).Error; err == nil {
					warehouse.AvailableArea += areaUsed
					if warehouse.AvailableArea > warehouse.TotalArea {
						warehouse.AvailableArea = warehouse.TotalArea
					}
					tx.Save(&warehouse)
				}

				// ✅ Profit math
				profit := (item.SellingPrice - entry.BillingPrice) * float64(qtyToOffboard)
				netProfit := profit - storageCost - avgExpense

				if err := tx.Create(&models.Profit{
					BatchID:   entry.BatchID,
					ProductID: entry.ProductID,
					Profit:    profit,
					NetProfit: netProfit,
				}).Error; err != nil {
					return fmt.Errorf("failed to record profit: %v", err)
				}

				// Add billing item
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

				// ✅ Mark batch inactive if empty
				var remaining int64
				tx.Model(&models.BatchProductEntry{}).
					Where("batch_id = ? AND stock_quantity > 0", entry.BatchID).
					Count(&remaining)
				if remaining == 0 {
					batch.Status = "inactive"
					tx.Save(&batch)
				}
			}

			if remainingQty > 0 {
				return fmt.Errorf("not enough stock to offboard %d units for product %v", item.OffboardQty, item.ProductID)
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

		if err := tx.Create(&billing).Error; err != nil {
			return fmt.Errorf("failed to create billing: %v", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ Transaction failed: %v", err)
	}
	return &billing, err
}

// GetByID fetches a single billing record with items by ID
func (r *BillingRepo) GetByID(ctx context.Context, id uint) (*models.Billing, error) {
	var billing models.Billing
	err := dbconn.DB.WithContext(ctx).
		Preload("Items").
		First(&billing, id).Error

	if err != nil {
		return nil, err
	}
	return &billing, nil
}

// GetAll fetches all billing records with their items
func (r *BillingRepo) GetAll(ctx context.Context) ([]models.Billing, error) {
	var billings []models.Billing
	err := dbconn.DB.WithContext(ctx).
		Preload("Items").
		Find(&billings).Error

	if err != nil {
		return nil, err
	}
	return billings, nil
}
