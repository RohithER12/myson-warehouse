package repo

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type AnalyticsRepo struct {
}

func NewAnalyticsRepo() *AnalyticsRepo {
	return &AnalyticsRepo{}
}

// 🔍 Get Analytics Data
func (r *AnalyticsRepo) GetAnalytics(ctx context.Context, warehouseID uint, duration string) (*models.ProductAnalytics, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var analytics models.ProductAnalytics

	// 🕒 Duration filter (used for flow metrics only)
	startDate := time.Now().AddDate(0, 0, -7) // default last 7 days
	switch duration {
	case "lastmonth":
		startDate = time.Now().AddDate(0, -1, 0)
	case "lastyear":
		startDate = time.Now().AddDate(-1, 0, 0)
	}

	// ===================================================
	// 📊 TOTAL AMOUNTS (flow metrics use startDate)
	// ===================================================
	db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Where("b.warehouse_id = ? AND be.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(be.billing_price * be.quantity), 0)").
		Scan(&analytics.TotalAmounts.OnBoardingAmount)

	db.Table(ns.TableName("BillingItem")+" AS bi").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON bi.batch_id = b.id").
		Where("b.warehouse_id = ? AND bi.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0)").
		Scan(&analytics.TotalAmounts.OffBoardingAmount)

	// InStockAmount for totals — use the snapshot of stock entries created/updated since startDate (keeps consistency with flow), but you can remove the created_at filter if you want absolute current stock value.
	db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Where("b.warehouse_id = ? AND be.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(be.billing_price * be.stock_quantity), 0)").
		Scan(&analytics.TotalAmounts.InStockAmount)

	db.Table(ns.TableName("Profit")+" AS p").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON p.batch_id = b.id").
		Where("b.warehouse_id = ? AND p.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(p.profit), 0)").
		Scan(&analytics.TotalAmounts.ProfitAmount)

	db.Table(ns.TableName("Profit")+" AS p").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON p.batch_id = b.id").
		Where("b.warehouse_id = ? AND p.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(p.net_profit), 0)").
		Scan(&analytics.TotalAmounts.NetProfitAmount)

	db.Table(ns.TableName("Billing")+" AS bl").
		Where("bl.created_at >= ?", startDate).
		Select("COALESCE(SUM(bl.other_expenses + bl.total_rent), 0)").
		Scan(&analytics.TotalAmounts.ExpenseAmount)

	// ===================================================
	// 🏭 GODOWN DATA
	// ===================================================
	var warehouse models.Warehouse
	if err := db.Preload("RentConfig").First(&warehouse, warehouseID).Error; err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	usedSpace := warehouse.TotalArea - warehouse.AvailableArea
	usedPercent := 0.0
	if warehouse.TotalArea > 0 {
		usedPercent = (usedSpace / warehouse.TotalArea) * 100
	}

	analytics.GodownData = models.GodownData{
		GodownID:            warehouse.ID,
		GodownName:          warehouse.Name,
		TotalSpace:          warehouse.TotalArea,
		AvailableSpace:      warehouse.AvailableArea,
		UsedSpace:           usedSpace,
		UsedSpacePercentage: usedPercent,
	}

	// ===================================================
	// 📦 PRODUCT-WISE ANALYTICS
	// ===================================================
	var products []models.Product
	// use Model+Preload so Supplier is loaded correctly
	if err := db.Model(&models.Product{}).Preload("Supplier").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	for _, p := range products {
		var pdata models.ProductWiseData

		// product info
		pdata.ProductInfo = models.ProductData{
			ID:          p.ID,
			Name:        p.Name,
			SupplierID:  p.SupplierID,
			Supplier:    p.Supplier,
			Category:    p.Category,
			StorageArea: p.StorageArea,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}

		// -----------------------
		// Amounts (flow metrics use startDate)
		// -----------------------
		db.Table(ns.TableName("BatchProductEntry")+" AS be").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ? AND be.created_at >= ?", warehouseID, p.ID, startDate).
			Select("COALESCE(SUM(be.billing_price * be.quantity), 0)").
			Scan(&pdata.Amounts.ProductOnBoardingAmount)

		db.Table(ns.TableName("BillingItem")+" AS bi").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON bi.batch_id = b.id").
			Where("b.warehouse_id = ? AND bi.product_id = ? AND bi.created_at >= ?", warehouseID, p.ID, startDate).
			Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0)").
			Scan(&pdata.Amounts.ProductOffBoardingAmount)

		// In-stock amount: use current stock snapshot (no created_at) so it reflects present inventory
		db.Table(ns.TableName("BatchProductEntry")+" AS be").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ? AND be.stock_quantity > 0", warehouseID, p.ID).
			Select("COALESCE(SUM(be.billing_price * be.stock_quantity), 0)").
			Scan(&pdata.Amounts.ProductInStockAmount)

		// Profit sums (use startDate for flow profit)
		var profitRes struct {
			Profit    float64
			NetProfit float64
		}
		db.Table(ns.TableName("Profit")+" AS pr").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON pr.batch_id = b.id").
			Where("b.warehouse_id = ? AND pr.product_id = ? AND pr.created_at >= ?", warehouseID, p.ID, startDate).
			Select("COALESCE(SUM(pr.profit),0) AS profit, COALESCE(SUM(pr.net_profit),0) AS net_profit").
			Scan(&profitRes)

		pdata.Amounts.ProductProfitAmount = profitRes.Profit
		pdata.Amounts.ProductNetProfitAmount = profitRes.NetProfit

		// Expense: storage_cost + shared other_expenses (use startDate to keep consistent)
		var productExpense float64
		db.Raw(fmt.Sprintf(`
			SELECT 
				COALESCE(SUM(bi.storage_cost), 0) +
				COALESCE(SUM(bl.other_expenses / NULLIF(prod_count.cnt, 0)), 0)
			FROM %s AS bi
			JOIN %s AS b ON bi.batch_id = b.id
			JOIN %s AS bl ON bi.billing_id = bl.id
			JOIN (
				SELECT billing_id, COUNT(DISTINCT product_id) AS cnt
				FROM %s
				GROUP BY billing_id
			) AS prod_count ON prod_count.billing_id = bl.id
			WHERE b.warehouse_id = ? AND bi.product_id = ? AND bl.created_at >= ?;
		`,
			ns.TableName("BillingItem"),
			ns.TableName("Batch"),
			ns.TableName("Billing"),
			ns.TableName("BillingItem")),
			warehouseID, p.ID, startDate).Scan(&productExpense)

		pdata.Amounts.ProductExpenseAmount = productExpense

		// -----------------------
		// Stock counts (current state)
		// -----------------------
		var stockRes struct {
			OnBoard  int
			InStock  int
			OffBoard int
		}

		db.Table(ns.TableName("BatchProductEntry")+" AS be").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(be.quantity),0) AS on_board, COALESCE(SUM(be.stock_quantity),0) AS in_stock").
			Scan(&stockRes)

		db.Table(ns.TableName("BillingItem")+" AS bi").
			Joins("JOIN "+ns.TableName("Batch")+" AS b ON bi.batch_id = b.id").
			Where("b.warehouse_id = ? AND bi.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(bi.offboard_qty),0) AS off_board").
			Scan(&stockRes.OffBoard)

		pdata.Stock.OnBoardCount = stockRes.OnBoard
		pdata.Stock.InStockCount = stockRes.InStock
		pdata.Stock.OffBoardCount = stockRes.OffBoard

		// -----------------------
		// Fast-moving flag (simple ratio)
		// -----------------------
		totalOn := float64(pdata.Stock.OnBoardCount)
		totalOff := float64(pdata.Stock.OffBoardCount)
		pdata.IsFastMoving = totalOn > 0 && (totalOff/totalOn) >= 0.7

		analytics.ProductsData = append(analytics.ProductsData, pdata)
	}

	// ===================================================
	// 🏆 TOP 10 PRODUCTS (by off/on ratio)
	// ===================================================
	type rankItem struct {
		Product models.ProductWiseData
		Score   float64
	}

	var ranks []rankItem
	for _, p := range analytics.ProductsData {
		on := float64(p.Stock.OnBoardCount)
		off := float64(p.Stock.OffBoardCount)
		if on <= 0 {
			// skip products with no onboarding (can't compute ratio sensibly)
			continue
		}
		score := off / on // higher ratio -> faster moving
		ranks = append(ranks, rankItem{Product: p, Score: score})
	}

	// sort descending
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Score > ranks[j].Score
	})

	// take top 10 (or fewer)
	limit := 10
	if len(ranks) < limit {
		limit = len(ranks)
	}
	analytics.TopTenProducts = make([]models.ProductCount, 0, limit)
	for i := 0; i < limit; i++ {
		analytics.TopTenProducts = append(analytics.TopTenProducts, models.ProductCount{
			ProductInfo: ranks[i].Product.ProductInfo,
			Stock:       ranks[i].Product.Stock,
		})
	}

	log.Printf("📈 Analytics generated for Warehouse %d (duration=%s) — products=%d top=%d",
		warehouseID, duration, len(analytics.ProductsData), len(analytics.TopTenProducts))

	return &analytics, nil
}

func (r *AnalyticsRepo) GetProductAnalyticsById(ctx context.Context, warehouseID, productID uint) (*models.ProductWiseAnalyticsData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	var pdata models.ProductWiseAnalyticsData

	// 🧱 Step 1: Get product info with supplier
	var product models.Product
	if err := db.Table(ns.TableName("Product")).
		Preload("Supplier").
		Where("id = ?", productID).
		First(&product).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	pdata.ProductInfo = models.ProductData{
		ID:          product.ID,
		Name:        product.Name,
		SupplierID:  product.SupplierID,
		Supplier:    product.Supplier,
		Category:    product.Category,
		StorageArea: product.StorageArea,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	// 🧮 Step 2: Amounts — OnBoarding, OffBoarding, InStock
	db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Where("be.product_id = ? AND b.warehouse_id = ?", productID, warehouseID).
		Select("COALESCE(SUM(be.billing_price * be.quantity), 0)").
		Scan(&pdata.Amounts.ProductOnBoardingAmount)

	db.Table(ns.TableName("BillingItem")+" AS bi").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON bi.batch_id = b.id").
		Where("bi.product_id = ? AND b.warehouse_id = ?", productID, warehouseID).
		Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0)").
		Scan(&pdata.Amounts.ProductOffBoardingAmount)

	db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Where("be.product_id = ? AND b.warehouse_id = ? AND be.stock_quantity > 0", productID, warehouseID).
		Select("COALESCE(SUM(be.billing_price * be.stock_quantity), 0)").
		Scan(&pdata.Amounts.ProductInStockAmount)

	// 🧮 Step 3: Profit + NetProfit
	var profitRes struct {
		Profit    float64
		NetProfit float64
	}
	db.Table(ns.TableName("Profit")+" AS pr").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON pr.batch_id = b.id").
		Where("pr.product_id = ? AND b.warehouse_id = ?", productID, warehouseID).
		Select("COALESCE(SUM(pr.profit),0) AS profit, COALESCE(SUM(pr.net_profit),0) AS net_profit").
		Scan(&profitRes)

	pdata.Amounts.ProductProfitAmount = profitRes.Profit
	pdata.Amounts.ProductNetProfitAmount = profitRes.NetProfit

	// 🧮 Step 4: Product Expense (storage_cost + shared other_expenses)
	var productExpense float64
	db.Raw(fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(bi.storage_cost), 0) +
			COALESCE(SUM(bl.other_expenses / NULLIF(prod_count.cnt, 0)), 0)
		FROM %s AS bi
		JOIN %s AS b ON bi.batch_id = b.id
		JOIN %s AS bl ON bi.billing_id = bl.id
		JOIN (
			SELECT billing_id, COUNT(DISTINCT product_id) AS cnt
			FROM %s
			GROUP BY billing_id
		) AS prod_count ON prod_count.billing_id = bl.id
		WHERE bi.product_id = ? AND b.warehouse_id = ?;
	`,
		ns.TableName("BillingItem"),
		ns.TableName("Batch"),
		ns.TableName("Billing"),
		ns.TableName("BillingItem")),
		productID, warehouseID).Scan(&productExpense)

	pdata.Amounts.ProductExpenseAmount = productExpense

	// 🧮 Step 5: Stock Counts
	var stockRes struct {
		OnBoard  int
		InStock  int
		OffBoard int
	}

	db.Table(ns.TableName("BatchProductEntry")+" AS be").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON be.batch_id = b.id").
		Where("b.warehouse_id = ? AND be.product_id = ?", warehouseID, productID).
		Select("COALESCE(SUM(be.quantity),0) AS on_board, COALESCE(SUM(be.stock_quantity),0) AS in_stock").
		Scan(&stockRes)

	db.Table(ns.TableName("BillingItem")+" AS bi").
		Joins("JOIN "+ns.TableName("Batch")+" AS b ON bi.batch_id = b.id").
		Where("b.warehouse_id = ? AND bi.product_id = ?", warehouseID, productID).
		Select("COALESCE(SUM(bi.offboard_qty),0) AS off_board").
		Scan(&stockRes.OffBoard)

	pdata.Stock.OnBoardCount = stockRes.OnBoard
	pdata.Stock.InStockCount = stockRes.InStock
	pdata.Stock.OffBoardCount = stockRes.OffBoard

	// ⚡ Step 6: Fast-moving check
	totalOn := float64(pdata.Stock.OnBoardCount)
	totalOff := float64(pdata.Stock.OffBoardCount)
	pdata.IsFastMoving = totalOn > 0 && (totalOff/totalOn) >= 0.7

	log.Printf("📦 Analytics fetched for Product %d in Warehouse %d", productID, warehouseID)
	return &pdata, nil
}

func (r *AnalyticsRepo) GetFastAndSlowMovingProductAnalytics(ctx context.Context, warehouseID uint) (map[string][]models.ProductWiseData, error) {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy

	// Load all products with suppliers
	var products []models.Product
	if err := db.Table(ns.TableName("Product")).Preload("Supplier").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// ----------------------------------------------
	// Get rent rate for this warehouse
	// ----------------------------------------------
	var rentRate float64
	db.Raw(`
		SELECT rr.rate_per_sqft
		FROM `+ns.TableName("Warehouse")+` w
		JOIN `+ns.TableName("RentRate")+` rr ON rr.id = w.rent_config_id
		WHERE w.id = ?
		LIMIT 1
	`, warehouseID).Scan(&rentRate)

	type candidate struct {
		Data         models.ProductWiseData
		RentPerSpace float64
		OffBoard     int
	}

	var candidates []candidate

	// Process each product
	for _, p := range products {

		var pdata models.ProductWiseData

		// Fill product info
		pdata.ProductInfo = models.ProductData{
			ID:          p.ID,
			Name:        p.Name,
			SupplierID:  p.SupplierID,
			Supplier:    p.Supplier,
			Category:    p.Category,
			StorageArea: p.StorageArea,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}

		// ----------------------------------------------
		// STOCK COUNT
		// ----------------------------------------------
		var stockRes struct {
			OnBoard  int
			InStock  int
			OffBoard int
		}

		db.Table(ns.TableName("BatchProductEntry")+" be").
			Joins("JOIN "+ns.TableName("Batch")+" b ON be.batch_id = b.id").
			Select("COALESCE(SUM(be.quantity),0) AS on_board, COALESCE(SUM(be.stock_quantity),0) AS in_stock").
			Where("be.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&stockRes)

		db.Table(ns.TableName("BillingItem")+" bi").
			Joins("JOIN "+ns.TableName("Batch")+" b ON bi.batch_id = b.id").
			Select("COALESCE(SUM(bi.offboard_qty),0)").
			Where("bi.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&stockRes.OffBoard)

		pdata.Stock.OnBoardCount = stockRes.OnBoard
		pdata.Stock.InStockCount = stockRes.InStock
		pdata.Stock.OffBoardCount = stockRes.OffBoard

		// Skip products with zero onboard count
		if pdata.Stock.OnBoardCount == 0 {
			continue
		}

		// ----------------------------------------------
		// AMOUNTS
		// ----------------------------------------------

		// Onboarding Amount
		db.Table(ns.TableName("BatchProductEntry")+" be").
			Joins("JOIN "+ns.TableName("Batch")+" b ON be.batch_id = b.id").
			Select("COALESCE(SUM(be.billing_price * be.quantity),0)").
			Where("be.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&pdata.Amounts.ProductOnBoardingAmount)

		// Offboarding Amount
		db.Table(ns.TableName("BillingItem")+" bi").
			Joins("JOIN "+ns.TableName("Batch")+" b ON bi.batch_id = b.id").
			Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty),0)").
			Where("bi.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&pdata.Amounts.ProductOffBoardingAmount)

		// Instock Amount
		db.Table(ns.TableName("BatchProductEntry")+" be").
			Joins("JOIN "+ns.TableName("Batch")+" b ON be.batch_id = b.id").
			Select("COALESCE(SUM(be.billing_price * be.stock_quantity),0)").
			Where("be.product_id = ? AND b.warehouse_id = ? AND be.stock_quantity > 0", p.ID, warehouseID).
			Scan(&pdata.Amounts.ProductInStockAmount)

		// Profit + Net Profit
		var profitRes struct {
			Profit    float64
			NetProfit float64
		}

		db.Table(ns.TableName("Profit")+" pr").
			Joins("JOIN "+ns.TableName("Batch")+" b ON pr.batch_id = b.id").
			Select("COALESCE(SUM(pr.profit),0) AS profit, COALESCE(SUM(pr.net_profit),0) AS net_profit").
			Where("pr.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&profitRes)

		pdata.Amounts.ProductProfitAmount = profitRes.Profit
		pdata.Amounts.ProductNetProfitAmount = profitRes.NetProfit

		// Expense (storage cost)
		db.Table(ns.TableName("BillingItem")+" bi").
			Joins("JOIN "+ns.TableName("Batch")+" b ON bi.batch_id = b.id").
			Select("COALESCE(SUM(bi.storage_cost),0)").
			Where("bi.product_id = ? AND b.warehouse_id = ?", p.ID, warehouseID).
			Scan(&pdata.Amounts.ProductExpenseAmount)

		// ----------------------------------------------
		// RENT PER SPACE
		// ----------------------------------------------
		rentPerUnit := rentRate * p.StorageArea
		offboardRent := pdata.Amounts.ProductExpenseAmount
		instockRent := rentPerUnit * float64(pdata.Stock.InStockCount)
		totalRent := offboardRent + instockRent

		rentPerSpace := totalRent / (float64(pdata.Stock.OnBoardCount) * p.StorageArea)
		pdata.Amounts.RentPerSpace = rentPerSpace

		candidates = append(candidates, candidate{
			Data:         pdata,
			RentPerSpace: rentPerSpace,
			OffBoard:     pdata.Stock.OffBoardCount,
		})
	}

	// ----------------------------------------------
	// SORT ASCENDING BY RentPerSpace
	// ----------------------------------------------
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].RentPerSpace < candidates[j].RentPerSpace
	})

	// ----------------------------------------------
	// FAST MOVING = lowest rentPerSpace AND offboard > 0
	// ----------------------------------------------
	var fast []models.ProductWiseData
	for _, c := range candidates {
		if c.OffBoard > 0 {
			fast = append(fast, c.Data)
			if len(fast) >= 5 {
				break
			}
		}
	}

	// ----------------------------------------------
	// SLOW MOVING = highest rentPerSpace (from end)
	// ----------------------------------------------
	var slow []models.ProductWiseData
	for i := len(candidates) - 1; i >= 0 && len(slow) < 5; i-- {
		slow = append(slow, candidates[i].Data)
	}

	// Set flags
	for i := range fast {
		fast[i].IsFastMoving = true
	}
	for i := range slow {
		slow[i].IsFastMoving = false
	}

	return map[string][]models.ProductWiseData{
		"fast_products": fast,
		"slow_products": slow,
	}, nil
}




