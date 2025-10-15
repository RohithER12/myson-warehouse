package repo

import (
	"context"
	"log"
	"math"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AnalyticsRepo struct {
	billingCol *mongo.Collection
	batchCol   *mongo.Collection
	productCol *mongo.Collection
}

func NewAnalyticsRepo() *AnalyticsRepo {
	return &AnalyticsRepo{
		billingCol: dbconn.GetCollection("myson_warehouse", "billings"),
		batchCol:   dbconn.GetCollection("myson_warehouse", "batches"),
		productCol: dbconn.GetCollection("myson_warehouse", "products"),
	}
}

type ProductAnalytics struct {
	ProductID      primitive.ObjectID `bson:"product_id" json:"product_id"`
	ProductName    string             `bson:"product_name" json:"product_name"`
	TotalStored    int                `bson:"total_stored" json:"total_stored"`
	TotalReleased  int                `bson:"total_released" json:"total_released"`
	AvgStorageTime float64            `bson:"avg_storage_time" json:"avg_storage_time"` // days
	TotalProfit    float64            `bson:"total_profit" json:"total_profit"`
	IsFastMoving   bool               `bson:"is_fast_moving" json:"is_fast_moving"`
}

func (r *AnalyticsRepo) GenerateProductAnalytics(ctx context.Context, productID primitive.ObjectID) (*ProductAnalytics, error) {
	// 1️⃣ Fetch product info
	var product models.Product
	if err := r.productCol.FindOne(ctx, bson.M{"_id": productID}).Decode(&product); err != nil {
		return nil, err
	}

	// 2️⃣ Fetch all batches for this product
	cursor, err := r.batchCol.Find(ctx, bson.M{"products.product_id": productID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var totalStored int
	for cursor.Next(ctx) {
		var batch models.Batch
		if err := cursor.Decode(&batch); err != nil {
			log.Println("Failed to decode batch:", err)
			continue
		}

		for _, p := range batch.Products {
			if p.ProductID == productID {
				totalStored += p.Quantity
			}
		}
	}

	// 3️⃣ Fetch all billing records for this product
	cursorBill, err := r.billingCol.Find(ctx, bson.M{"product_id": productID.Hex()})
	if err != nil {
		return nil, err
	}
	defer cursorBill.Close(ctx)

	var totalReleased int
	var totalProfit float64
	var totalDays float64
	var billingCount int

	for cursorBill.Next(ctx) {
		var b models.Billing
		if err := cursorBill.Decode(&b); err != nil {
			log.Println("Failed to decode billing:", err)
			continue
		}

		for _, ub := range b.BatchesUsed {
			totalReleased += ub.Quantity
		}

		totalProfit += b.Margin
		totalDays += b.StorageDuration
		billingCount++
	}

	avgStorageTime := 0.0
	if billingCount > 0 {
		avgStorageTime = totalDays / float64(billingCount)
	}

	// 4️⃣ Determine if product is fast moving
	isFastMoving := false
	if avgStorageTime < 5 {
		isFastMoving = true
	} else {
		// last 30 days
		cutoff := time.Now().AddDate(0, 0, -30)
		releasedLast30 := 0

		cursor30, err := r.billingCol.Find(ctx, bson.M{"product_id": productID.Hex(), "start_date": bson.M{"$gte": cutoff}})
		if err == nil {
			defer cursor30.Close(ctx)
			for cursor30.Next(ctx) {
				var b models.Billing
				if err := cursor30.Decode(&b); err == nil {
					for _, ub := range b.BatchesUsed {
						releasedLast30 += ub.Quantity
					}
				}
			}
		}
		if totalReleased > 0 && float64(releasedLast30)/float64(totalReleased) > 0.5 {
			isFastMoving = true
		}
	}

	analytics := &ProductAnalytics{
		ProductID:      product.ID,
		ProductName:    product.Name,
		TotalStored:    totalStored,
		TotalReleased:  totalReleased,
		AvgStorageTime: math.Round(avgStorageTime*100) / 100,
		TotalProfit:    totalProfit,
		IsFastMoving:   isFastMoving,
	}

	return analytics, nil
}

// Generate analytics for all products
func (r *AnalyticsRepo) GenerateAllProductsAnalytics(ctx context.Context) ([]*ProductAnalytics, error) {
	cursor, err := r.productCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var analyticsList []*ProductAnalytics
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			log.Println("Failed to decode product:", err)
			continue
		}

		analytics, err := r.GenerateProductAnalytics(ctx, product.ID)
		if err != nil {
			log.Println("Failed analytics for product:", product.Name, err)
			continue
		}
		analyticsList = append(analyticsList, analytics)
	}

	return analyticsList, nil
}
