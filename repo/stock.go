package repo

import (
	"context"
	"log"
	"strings"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type ProductStockRepo struct {
}

// NewStockRepo initializes the repository
func NewProductStockRepo() *ProductStockRepo {
	return &ProductStockRepo{}
}

func (r *ProductStockRepo) GetProductStockWithRent(ctx context.Context) ([]map[string]any, error) {
	var entries []models.ProductStockView

	err := dbconn.DB.WithContext(ctx).
		Table("batches AS b").
		Select(`
			b.id AS batch_id,
			b.warehouse_id,
			w.name AS warehouse_name,
			p.id AS product_id,
			p.name AS product_name,
			p.category,
			s.name AS supplier_name,
			p.storage_area,
			be.quantity,
			be.stock_quantity,
			be.billing_price,
			b.created_at,
			be.last_updated,
			rr.rate_per_sqft,
			rr.currency,
			rr.billing_cycle
		`).
		Joins("JOIN batch_product_entries AS be ON b.id = be.batch_id").
		Joins("JOIN products AS p ON be.product_id = p.id").
		Joins("JOIN suppliers AS s ON p.supplier_id = s.id").
		Joins("JOIN warehouses AS w ON b.warehouse_id = w.id").
		Joins("JOIN rent_rates AS rr ON w.rent_config_id = rr.id").
		// Where("b.deleted_at IS NULL AND be.deleted_at IS NULL").
		Find(&entries).Error

	if err != nil {
		log.Printf("❌ Query error: %v\n", err)
		return nil, err
	}

	now := time.Now()
	var results []map[string]interface{}

	for _, e := range entries {
		daysStored := now.Sub(e.CreatedAt).Hours() / 24
		if daysStored < 1 {
			daysStored = 1
		}

		var rent float64
		switch strings.ToLower(e.BillingCycle) {
		case "daily":
			rent = daysStored * e.StorageArea * e.RatePerSqft * float64(e.StockQuantity)
		case "monthly":
			rent = (daysStored / 30) * e.StorageArea * e.RatePerSqft * float64(e.StockQuantity)
		default:
			rent = (daysStored / 30) * e.StorageArea * e.RatePerSqft * float64(e.StockQuantity)
		}

		status := "out_of_stock"
		if e.StockQuantity > 0 {
			status = "in_stock"
		}

		results = append(results, map[string]any{
			"batch_id":       e.BatchID,
			"product_id":     e.ProductID,
			"product_name":   e.ProductName,
			"category":       e.Category,
			"supplier_name":  e.SupplierName,
			"warehouse_id":   e.WarehouseID,
			"warehouse_name": e.WarehouseName,
			"billing_price":  e.BillingPrice,
			"batch_quantity": e.Quantity,
			"stock_quantity": e.StockQuantity,
			"stored_at":      e.CreatedAt,
			"last_updated":   e.LastUpdated,
			"total_space":    e.StorageArea * float64(e.StockQuantity),
			"total_rent":     rent,
			"currency":       e.Currency,
			"status":         status,
		})
	}

	log.Printf("🏁 Final Result count: %d entries", len(results))
	return results, nil
}

// GetAllproducts aggregates product stock and related details
func (r *ProductStockRepo) GetAllproducts(ctx context.Context) ([]models.BasicProductStockView, error) {
	var results []models.BasicProductStockView
	log.Println("herer")
	err := dbconn.DB.WithContext(ctx).
		Table("batch_product_entries AS bpe").
		Select(`
			b.warehouse_id,
			w.name AS warehouse_name,
			p.id AS product_id,
			p.name AS product_name,
			p.category,
			AVG(p.storage_area) AS average_storage_area,
			SUM(bpe.stock_quantity) AS stock_quantity,
			AVG(bpe.billing_price) AS average_billing_price,
			AVG(rr.rate_per_sqft) AS average_rate_per_sqft,
			rr.currency,
			rr.billing_cycle
		`).
		Joins("JOIN batches b ON b.id = bpe.batch_id").
		Joins("JOIN products p ON p.id = bpe.product_id").
		Joins("JOIN warehouses w ON w.id = b.warehouse_id").
		Joins("JOIN rent_rates rr ON rr.id = w.rent_config_id").
		Group("b.warehouse_id, w.name, p.id, p.name, p.category, p.storage_area, rr.currency, rr.billing_cycle").
		Order("w.name, p.name").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
