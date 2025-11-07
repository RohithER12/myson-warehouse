package repo

import (
	"context"
	"fmt"
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
	var analytics models.ProductAnalytics

	// 🕒 Duration filter
	startDate := time.Now().AddDate(0, 0, -7) // Default last 7 days
	switch duration {
	case "lastmonth":
		startDate = time.Now().AddDate(0, -1, 0)
	case "lastyear":
		startDate = time.Now().AddDate(-1, 0, 0)
	}

	// =============================
	// 📊 TOTAL AMOUNTS SECTION
	// =============================

	// OnBoardingAmount
	db.Table("batch_product_entries AS be").
		Joins("JOIN batches AS b ON be.batch_id = b.id").
		Where("b.warehouse_id = ? AND be.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(be.billing_price * be.quantity), 0)").
		Scan(&analytics.TotalAmounts.OnBoardingAmount)

	// OffBoardingAmount
	db.Table("billing_items AS bi").
		Joins("JOIN batches AS b ON bi.batch_id = b.id").
		Where("b.warehouse_id = ? AND bi.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0)").
		Scan(&analytics.TotalAmounts.OffBoardingAmount)

	// InStockAmount
	db.Table("batch_product_entries AS be").
		Joins("JOIN batches AS b ON be.batch_id = b.id").
		Where("b.warehouse_id = ? AND be.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(be.billing_price * be.stock_quantity), 0)").
		Scan(&analytics.TotalAmounts.InStockAmount)

	// ProfitAmount
	db.Table("profits AS p").
		Joins("JOIN batches AS b ON p.batch_id = b.id").
		Where("b.warehouse_id = ? AND p.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(p.profit), 0)").
		Scan(&analytics.TotalAmounts.ProfitAmount)

	// NetProfitAmount
	db.Table("profits AS p").
		Joins("JOIN batches AS b ON p.batch_id = b.id").
		Where("b.warehouse_id = ? AND p.created_at >= ?", warehouseID, startDate).
		Select("COALESCE(SUM(p.net_profit), 0)").
		Scan(&analytics.TotalAmounts.NetProfitAmount)

	// ExpenseAmount
	db.Table("billings AS bi").
		Where("bi.created_at >= ?", startDate).
		Select("COALESCE(SUM(bi.other_expenses + bi.total_rent), 0)").
		Scan(&analytics.TotalAmounts.ExpenseAmount)

	// =============================
	// 🏭 GODOWN DATA SECTION
	// =============================
	var warehouse models.Warehouse
	if err := db.Preload("RentConfig").First(&warehouse, warehouseID).Error; err != nil {
		return nil, fmt.Errorf("warehouse not found: %v", err)
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

	// =============================
	// 📦 PRODUCT-WISE ANALYTICS
	// =============================
	var products []models.Product
	if err := db.Preload("Supplier").Find(&products).Error; err != nil {
		return nil, err
	}

	for _, p := range products {
		var pdata models.ProductWiseData
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

		// =============================
		// 💰 PRODUCT AMOUNTS
		// =============================

		// Onboarding = ∑(billing_price * quantity)
		db.Table("batch_product_entries AS be").
			Joins("JOIN batches AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(be.billing_price * be.quantity), 0)").
			Scan(&pdata.Amounts.ProductOnBoardingAmount)

		// Offboarding = ∑(selling_price * offboard_qty)
		db.Table("billing_items AS bi").
			Joins("JOIN batches AS b ON bi.batch_id = b.id").
			Where("b.warehouse_id = ? AND bi.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(bi.selling_price * bi.offboard_qty), 0)").
			Scan(&pdata.Amounts.ProductOffBoardingAmount)

		// In stock = ∑(billing_price * stock_quantity)
		db.Table("batch_product_entries AS be").
			Joins("JOIN batches AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ? AND be.stock_quantity > 0", warehouseID, p.ID).
			Select("COALESCE(SUM(be.billing_price * be.stock_quantity), 0)").
			Scan(&pdata.Amounts.ProductInStockAmount)

		// Profit + Net Profit (direct from profits table)
		var profitRes struct {
			TotalProfit    float64
			TotalNetProfit float64
		}
		db.Table("profits AS pr").
			Joins("JOIN batches AS b ON pr.batch_id = b.id").
			Where("b.warehouse_id = ? AND pr.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(pr.profit), 0) AS total_profit, COALESCE(SUM(pr.net_profit), 0) AS total_net_profit").
			Scan(&profitRes)

		pdata.Amounts.ProductProfitAmount = profitRes.TotalProfit
		pdata.Amounts.ProductNetProfitAmount = profitRes.TotalNetProfit

		// Expense = ∑(storage_cost) + proportional share of other_expenses
		var productExpense float64
		db.Raw(`
			SELECT 
				COALESCE(SUM(bi.storage_cost), 0) +
				COALESCE(SUM(bl.other_expenses / NULLIF(prod_count.cnt, 0)), 0)
			FROM billing_items bi
			JOIN batches b ON bi.batch_id = b.id
			JOIN billings bl ON bi.billing_id = bl.id
			JOIN (
				SELECT billing_id, COUNT(DISTINCT product_id) AS cnt
				FROM billing_items
				GROUP BY billing_id
			) AS prod_count ON prod_count.billing_id = bl.id
			WHERE b.warehouse_id = ? AND bi.product_id = ?;
		`, warehouseID, p.ID).Scan(&productExpense)
		pdata.Amounts.ProductExpenseAmount = productExpense

		// =============================
		// 📦 STOCK COUNTS
		// =============================
		var stockRes struct {
			OnBoard  int
			InStock  int
			OffBoard int
		}

		db.Table("batch_product_entries AS be").
			Joins("JOIN batches AS b ON be.batch_id = b.id").
			Where("b.warehouse_id = ? AND be.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(be.quantity),0) AS on_board, COALESCE(SUM(be.stock_quantity),0) AS in_stock").
			Scan(&stockRes)

		db.Table("billing_items AS bi").
			Joins("JOIN batches AS b ON bi.batch_id = b.id").
			Where("b.warehouse_id = ? AND bi.product_id = ?", warehouseID, p.ID).
			Select("COALESCE(SUM(bi.offboard_qty),0) AS off_board").
			Scan(&stockRes.OffBoard)

		pdata.Stock.OnBoardCount = stockRes.OnBoard
		pdata.Stock.InStockCount = stockRes.InStock
		pdata.Stock.OffBoardCount = stockRes.OffBoard

		// ⚡ Fast-moving logic
		totalOn := float64(pdata.Stock.OnBoardCount)
		totalOff := float64(pdata.Stock.OffBoardCount)
		pdata.IsFastMoving = totalOn > 0 && (totalOff/totalOn) >= 0.7

		analytics.ProductsData = append(analytics.ProductsData, pdata)
	}

	return &analytics, nil
}




