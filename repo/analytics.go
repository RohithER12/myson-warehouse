package repo

import (
	"context"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type AnalyticsRepo struct {
}

func NewAnalyticsRepo() *AnalyticsRepo {
	return &AnalyticsRepo{}
}

// 🔍 Get batch by ID
func (r *AnalyticsRepo) GetAnalytics(ctx context.Context, duration string) (*models.ProductAnalytics, error) {

	db := dbconn.DB.WithContext(ctx)

	analytics := &models.ProductAnalytics{}

	// Apply duration filter
	var startDate time.Time
	now := time.Now()
	switch duration {
	case "lastweek":
		startDate = now.AddDate(0, 0, -7)
	case "lastmonth":
		startDate = now.AddDate(0, -1, 0)
	case "lastyear":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // default last month
	}

	// --- Step 1: Fetch Data ---
	var batches []models.BatchProductEntry
	var billings []models.Billing
	var billingItems []models.BillingItem
	var warehouses []models.Warehouse

	db.Preload("Product").Find(&batches)
	db.Preload("RentConfig").Find(&warehouses)
	db.Preload("Items").Find(&billings, "created_at >= ?", startDate)
	db.Find(&billingItems, "created_at >= ?", startDate)

	// --- Step 2: Initialize totals ---
	var totalOnBoard, totalOffBoard, totalInStock, totalProfit, totalExpenses, totalNetProfit float64

	// --- Step 3: Per-Product Aggregation Map ---
	productMap := map[uint]*models.ProductWiseData{}

	// --- Step 4: OnBoard + InStock ---
	for _, b := range batches {
		p, ok := productMap[b.ProductID]
		if !ok {
			p = &models.ProductWiseData{}
			productMap[b.ProductID] = p
		}

		// Onboard amount
		p.Amounts.OnBoardingAmount += b.BillingPrice * float64(b.Quantity)
		totalOnBoard += b.BillingPrice * float64(b.Quantity)

		// In-stock amount
		p.Amounts.InStockAmount += b.BillingPrice * float64(b.StockQuantity)
		totalInStock += b.BillingPrice * float64(b.StockQuantity)

		// Stock counts
		p.Stock.InStockCount += b.StockQuantity
		p.Stock.OnBoardCount += b.Quantity
	}

	// --- Step 5: Offboard + Profit + Expenses ---
	for _, item := range billingItems {
		p, ok := productMap[item.ProductID]
		if !ok {
			p = &models.ProductWiseData{}
			productMap[item.ProductID] = p
		}

		// Offboarded amount
		offBoardAmount := item.SellingPrice * float64(item.OffboardQty)
		p.Amounts.OffBoardingAmount += offBoardAmount
		totalOffBoard += offBoardAmount

		// Profit amount
		profit := (item.SellingPrice - item.BuyingPrice) * float64(item.OffboardQty)
		p.Amounts.ProfitAmount += profit
		totalProfit += profit

		// Expense from storage + billing’s other expenses
		p.Amounts.ExpenseAmount += item.StorageCost
		totalExpenses += item.StorageCost

		// Stock offboard count
		p.Stock.OffBoardCount += item.OffboardQty
	}

	// --- Step 6: Add global billing other expenses ---
	var totalOtherExpense float64
	for _, b := range billings {
		totalOtherExpense += b.OtherExpenses
	}
	totalExpenses += totalOtherExpense

	// --- Step 7: Compute Fast Moving & Net Profit ---
	for _, p := range productMap {
		p.Amounts.NetProfitAmount = p.Amounts.ProfitAmount - p.Amounts.ExpenseAmount

		totalNetProfit += p.Amounts.NetProfitAmount

		totalCount := p.Stock.OnBoardCount
		if totalCount > 0 {
			moveRatio := float64(p.Stock.OffBoardCount) / float64(totalCount)
			p.IsFastMoving = moveRatio > 0.6
		}
	}

	// --- Step 8: Godown data (single example warehouse) ---
	if len(warehouses) > 0 {
		w := warehouses[0]
		usedSpace := w.TotalArea - w.AvailableArea
		analytics.GodownData = models.GodownData{
			GodownID:            w.ID,
			GodownName:          w.Name,
			TotalSpace:          w.TotalArea,
			AvailableSpace:      w.AvailableArea,
			UsedSpace:           usedSpace,
			UsedSpacePercentage: (usedSpace / w.TotalArea) * 100,
		}
	}

	// --- Step 9: Aggregate totals ---
	analytics.TotalAmounts = models.TotalAmounts{
		OnBoardingAmount:  totalOnBoard,
		OffBoardingAmount: totalOffBoard,
		InStockAmount:     totalInStock,
		ProfitAmount:      totalProfit,
		ExpenseAmount:     totalExpenses,
		NetProfitAmount:   totalProfit - totalExpenses,
	}

	for _, p := range productMap {
		analytics.ProductsData = append(analytics.ProductsData, *p)
	}

	return analytics, nil
}
