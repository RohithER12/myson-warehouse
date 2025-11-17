package repo

import (
	"context"
	"fmt"
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

	db := dbconn.DB.WithContext(ctx)
	ns := dbconn.DB.NamingStrategy // ✅ use the same naming strategy globally

	// ✅ Dynamically resolve prefixed table names
	batchTable := ns.TableName("Batch")
	beTable := ns.TableName("BatchProductEntry")
	productTable := ns.TableName("Product")
	supplierTable := ns.TableName("Supplier")
	warehouseTable := ns.TableName("Warehouse")
	rentRateTable := ns.TableName("RentRate")

	// ✅ Use dynamic names in query
	err := db.Table(fmt.Sprintf("%s AS b", batchTable)).
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
		Joins(fmt.Sprintf("JOIN %s AS be ON b.id = be.batch_id", beTable)).
		Joins(fmt.Sprintf("JOIN %s AS p ON be.product_id = p.id", productTable)).
		Joins(fmt.Sprintf("JOIN %s AS s ON p.supplier_id = s.id", supplierTable)).
		Joins(fmt.Sprintf("JOIN %s AS w ON b.warehouse_id = w.id", warehouseTable)).
		Joins(fmt.Sprintf("JOIN %s AS rr ON w.rent_config_id = rr.id", rentRateTable)).
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
	db := dbconn.DB.WithContext(ctx)
	ns := dbconn.DB.NamingStrategy // ✅ Use GORM's naming strategy

	// ✅ Dynamically resolve actual table names based on the prefix
	batchTable := ns.TableName("Batch")           // mys_batch
	bpeTable := ns.TableName("BatchProductEntry") // mys_batch_product_entry
	productTable := ns.TableName("Product")       // mys_product
	warehouseTable := ns.TableName("Warehouse")   // mys_warehouse
	rentRateTable := ns.TableName("RentRate")     // mys_rent_rate

	log.Println("Fetching product stock data (prefix-safe) ...")

	// ✅ Use dynamic table names in your joins
	err := db.Table(fmt.Sprintf("%s AS bpe", bpeTable)).
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
		Joins(fmt.Sprintf("JOIN %s AS b ON b.id = bpe.batch_id", batchTable)).
		Joins(fmt.Sprintf("JOIN %s AS p ON p.id = bpe.product_id", productTable)).
		Joins(fmt.Sprintf("JOIN %s AS w ON w.id = b.warehouse_id", warehouseTable)).
		Joins(fmt.Sprintf("JOIN %s AS rr ON rr.id = w.rent_config_id", rentRateTable)).
		Group("b.warehouse_id, w.name, p.id, p.name, p.category, p.storage_area, rr.currency, rr.billing_cycle").
		Order("w.name, p.name").
		Scan(&results).Error

	if err != nil {
		log.Printf("❌ Error fetching product stock: %v", err)
		return nil, err
	}

	log.Printf("✅ Retrieved %d product stock entries", len(results))
	return results, nil
}

func (r *ProductStockRepo) GetAllStockProductData(ctx context.Context) ([]models.StockSearchData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	type rawRow struct {
		ProductID      uint
		ProductName    string
		SupplierName   string
		Category       string
		StorageArea    float64
		WarehouseID    uint
		WarehouseName  string
		BatchID        uint
		OnBoardCount   int
		OffBoardCount  int
		InStockCount   int
		OnBoardingAmt  float64
		OffBoardingAmt float64
		InStockAmt     float64
		ProfitAmt      float64
		NetProfitAmt   float64
		ExpenseAmt     float64
	}

	var rows []rawRow

	// 🧮 Query core data per product-batch
	err := db.Table(ns.TableName("BatchProductEntry") + " AS be").
		Select(`
			p.id AS product_id,
			p.name AS product_name,
			s.name AS supplier_name,
			p.category,
			p.storage_area,
			b.warehouse_id,
			w.name AS warehouse_name,
			be.batch_id,
			COALESCE(SUM(be.quantity), 0) AS on_board_count,
			COALESCE(SUM(be.stock_quantity), 0) AS in_stock_count,
			COALESCE(SUM(bi.offboard_qty), 0) AS off_board_count,
			COALESCE(SUM(be.billing_price * be.quantity), 0) AS on_boarding_amt,
			COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0) AS off_boarding_amt,
			COALESCE(SUM(be.billing_price * be.stock_quantity), 0) AS in_stock_amt,
			COALESCE(SUM(pf.profit), 0) AS profit_amt,
			COALESCE(SUM(pf.net_profit), 0) AS net_profit_amt,
			COALESCE(SUM(bl.other_expenses / NULLIF(prod_count.cnt, 1)), 0) AS expense_amt
		`).
		Joins("JOIN " + ns.TableName("Product") + " AS p ON be.product_id = p.id").
		Joins("JOIN " + ns.TableName("Supplier") + " AS s ON p.supplier_id = s.id").
		Joins("JOIN " + ns.TableName("Batch") + " AS b ON be.batch_id = b.id").
		Joins("JOIN " + ns.TableName("Warehouse") + " AS w ON b.warehouse_id = w.id").
		Joins("LEFT JOIN " + ns.TableName("BillingItem") + " AS bi ON bi.batch_id = b.id AND bi.product_id = p.id").
		Joins("LEFT JOIN " + ns.TableName("Billing") + " AS bl ON bl.id = bi.billing_id").
		Joins("LEFT JOIN " + ns.TableName("Profit") + " AS pf ON pf.product_id = p.id AND pf.batch_id = b.id").
		Joins(`
			LEFT JOIN (
				SELECT billing_id, COUNT(DISTINCT product_id) AS cnt
				FROM ` + ns.TableName("BillingItem") + `
				GROUP BY billing_id
			) AS prod_count ON prod_count.billing_id = bl.id
		`).
		Group(`
			p.id, p.name, s.name, p.category, p.storage_area,
			b.warehouse_id, w.name, be.batch_id
		`).
		Order("p.name ASC, b.created_at ASC").
		Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch stock product data: %w", err)
	}

	if len(rows) == 0 {
		log.Printf("⚠️ No product stock data found")
		return nil, nil
	}

	// 🧩 Group data by product for structured output
	productMap := make(map[uint]*models.StockSearchData)

	for _, r := range rows {
		if _, exists := productMap[r.ProductID]; !exists {
			productMap[r.ProductID] = &models.StockSearchData{
				ProductID:     r.ProductID,
				ProductName:   r.ProductName,
				SupplierName:  r.SupplierName,
				Category:      r.Category,
				StorageArea:   r.StorageArea,
				WarehouseID:   r.WarehouseID,
				WarehouseName: r.WarehouseName,
			}
		}

		// Add per-batch stock info
		stockData := models.StockData{
			BatchID: r.BatchID,
			StockCount: models.Stock{
				OnBoardCount:  r.OnBoardCount,
				OffBoardCount: r.OffBoardCount,
				InStockCount:  r.InStockCount,
			},
			Amounts: models.TotalAmounts{
				OnBoardingAmount:  r.OnBoardingAmt,
				OffBoardingAmount: r.OffBoardingAmt,
				InStockAmount:     r.InStockAmt,
				ProfitAmount:      r.ProfitAmt,
				NetProfitAmount:   r.NetProfitAmt,
				ExpenseAmount:     r.ExpenseAmt,
			},
		}

		productMap[r.ProductID].StockData = append(productMap[r.ProductID].StockData, stockData)
	}

	// 🧾 Convert map to slice
	var result []models.StockSearchData
	for _, v := range productMap {
		result = append(result, *v)
	}

	log.Printf("📦 Retrieved %d products with stock data", len(result))
	return result, nil
}

func (r *ProductStockRepo) GetStockProductData(ctx context.Context, productId uint) (models.StockSearchData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	type rawRow struct {
		ProductID      uint
        ProductName    string
        SupplierName   string
        Category       string
        StorageArea    float64
        WarehouseID    uint
        WarehouseName  string
        RentPerSqft    float64
        BatchID        uint
        BatchCreatedAt time.Time

        OnBoardCount   int
        OffBoardCount  int
        InStockCount   int

        OnBoardingAmt  float64
        OffBoardingAmt float64
        InStockAmt     float64
        ProfitAmt      float64
        NetProfitAmt   float64
        ExpenseAmt     float64
	}

	var rows []rawRow

	// -------------------------------------------------------------
	// 🧠 Query all batches of a single product with aggregated info
	// -------------------------------------------------------------
	err := db.Table(ns.TableName("BatchProductEntry") + " AS be").
		Select(`
		    p.id AS product_id,
		    p.name AS product_name,
		    s.name AS supplier_name,
		    p.category,
		    p.storage_area,

		    b.warehouse_id,
		    w.name AS warehouse_name,
		    rr.rate_per_sqft AS rent_per_sqft,

		    be.batch_id,
		    b.created_at AS batch_created_at,

		    COALESCE(SUM(be.quantity), 0) AS on_board_count,
		    COALESCE(SUM(be.stock_quantity), 0) AS in_stock_count,
		    COALESCE(SUM(bi.offboard_qty), 0) AS off_board_count,

		    COALESCE(SUM(be.billing_price * be.quantity), 0) AS on_boarding_amt,
		    COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0) AS off_boarding_amt,
		    COALESCE(SUM(be.billing_price * be.stock_quantity), 0) AS in_stock_amt,

		    COALESCE(SUM(pf.profit), 0) AS profit_amt,
		    COALESCE(SUM(pf.net_profit), 0) AS net_profit_amt,

		    COALESCE(SUM(bl.other_expenses / NULLIF(prod_count.cnt, 1)), 0) AS expense_amt
		`).
		Joins("JOIN "+ns.TableName("Product")+" AS p ON be.product_id = p.id").
		Joins("JOIN "+ns.TableName("Supplier")+" AS s ON p.supplier_id = s.id").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Joins("JOIN "+ns.TableName("Warehouse")+" AS w ON b.warehouse_id = w.id").
		Joins("JOIN "+ns.TableName("RentRate")+" AS rr ON w.rent_config_id = rr.id").
		Joins("LEFT JOIN "+ns.TableName("BillingItem")+" AS bi ON bi.batch_id = b.id AND bi.product_id = p.id").
		Joins("LEFT JOIN "+ns.TableName("Billing")+" AS bl ON bl.id = bi.billing_id").
		Joins("LEFT JOIN "+ns.TableName("Profit")+" AS pf ON pf.product_id = p.id AND pf.batch_id = b.id").
		Joins(`
			LEFT JOIN (
			    SELECT billing_id, COUNT(DISTINCT product_id) AS cnt
			    FROM ` + ns.TableName("BillingItem") + `
			    GROUP BY billing_id
			) AS prod_count ON prod_count.billing_id = bl.id
		`).
		Where("p.id = ?", productId).
		Group(`
			p.id, p.name, s.name, p.category, p.storage_area,
			b.warehouse_id, w.name, rr.rate_per_sqft,
			be.batch_id, b.created_at
		`).
		Order("be.batch_id ASC").
		Scan(&rows).Error

	if err != nil {
		return models.StockSearchData{}, fmt.Errorf("failed to fetch stock data for product: %w", err)
	}

	if len(rows) == 0 {
		return models.StockSearchData{}, fmt.Errorf("no stock data found for product ID %d", productId)
	}

	// -------------------------------------------------------------
	// 🧱 Build response base (product details)
	// -------------------------------------------------------------
	head := rows[0]
	result := models.StockSearchData{
		ProductID:     head.ProductID,
		ProductName:   head.ProductName,
		SupplierName:  head.SupplierName,
		Category:      head.Category,
		StorageArea:   head.StorageArea,
		WarehouseID:   head.WarehouseID,
		WarehouseName: head.WarehouseName,
	}

	// -------------------------------------------------------------
	// 📦 Loop each batch and compute Rent + filter zero-stock
	// -------------------------------------------------------------
	for _, r := range rows {

		// ❌ Skip batches with 0 stock
		if r.InStockCount == 0 {
			continue
		}

		// 📅 Duration in days
		duration := int(time.Since(r.BatchCreatedAt).Hours() / 24)
		if duration < 1 {
			duration = 1
		}

		// 💰 RentAmount = storage_area * rent_per_sqft * duration * quantity
		rentAmount := r.StorageArea * r.RentPerSqft * float64(duration) * float64(r.InStockCount)

		// Add structured batch entry
		result.StockData = append(result.StockData, models.StockData{
			BatchID: r.BatchID,
			StockCount: models.Stock{
				OnBoardCount:  r.OnBoardCount,
				OffBoardCount: r.OffBoardCount,
				InStockCount:  r.InStockCount,
			},
			Amounts: models.TotalAmounts{
				OnBoardingAmount:  r.OnBoardingAmt,
				OffBoardingAmount: r.OffBoardingAmt,
				InStockAmount:     r.InStockAmt,
				ProfitAmount:      r.ProfitAmt,
				NetProfitAmount:   r.NetProfitAmt,
				ExpenseAmount:     r.ExpenseAmt,
			},
			RentAmount: rentAmount,
		})
	}

	log.Printf("📦 Retrieved stock data for product %d with %d batches", productId, len(result.StockData))
	return result, nil
}

