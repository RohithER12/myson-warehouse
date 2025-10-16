package repo

import (
	"context"
	"fmt"
	"log"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const billingCollection = "billings"

type BillingRepo struct {
	col         *mongo.Collection
	productRepo *ProductRepo
}

// NewBillingRepo initializes the billing repo
func NewBillingRepo() *BillingRepo {
	return &BillingRepo{
		col:         dbconn.GetCollection("myson_warehouse", billingCollection),
		productRepo: NewProductRepo(),
	}
}

type BillingItemInput struct {
	ProductID    string  `json:"product_id"`
	BatchID      string  `json:"batch_id"`
	OffboardQty  int     `json:"offboard_quantity"`
	SellingPrice float64 `json:"selling_price"`
}

func (r *BillingRepo) GenerateBilling(
	ctx context.Context,
	items []BillingItemInput,
	endDate time.Time,
	rentPerUnit float64,
	expenses []models.Expense,
) (*models.Billing, error) {

	productRepo := NewProductRepo()
	stockRepo := NewStockRepo()

	var billingItems []models.BillingItem
	var totalStorage, totalBuying, totalSelling float64

	for _, item := range items {
		// Fetch batch
		batch, err := stockRepo.GetBatchByID(ctx, item.BatchID)
		if err != nil {
			return nil, fmt.Errorf("batch %s not found: %v", item.BatchID, err)
		}

		// Find product entry inside batch
		var batchProduct *models.BatchProductEntry
		for i := range batch.Products {
			if batch.Products[i].ProductID.Hex() == item.ProductID {
				batchProduct = &batch.Products[i]
				break
			}
		}
		if batchProduct == nil {
			return nil, fmt.Errorf("product %s not found in batch %s", item.ProductID, item.BatchID)
		}
		if batchProduct.Quantity < item.OffboardQty {
			return nil, fmt.Errorf("not enough quantity in batch %s for product %s", item.BatchID, item.ProductID)
		}

		// Fetch product for storage area
		product, err := productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product fetch failed: %v", err)
		}

		// Duration from batch StoredAt to billing end date
		durationDays := endDate.Sub(batch.StoredAt).Hours() / 24
		if durationDays < 1 {
			durationDays = 1
		}

		storageCost := rentPerUnit * product.StorageArea * float64(item.OffboardQty) * durationDays
		totalBuyingPrice := batchProduct.BillingPrice * float64(item.OffboardQty)
		totalSellingPrice := item.SellingPrice * float64(item.OffboardQty)

		batchProduct.Quantity -= item.OffboardQty
		status := "partially_offboarded"
		if batchProduct.Quantity == 0 {
			status = "fully_offboarded"
		}
		batch.Status = status

		if err := stockRepo.UpdateBatch(ctx, batch); err != nil {
			log.Println("⚠️ Batch update failed:", err)
		}

		billingItems = append(billingItems, models.BillingItem{
			ProductID:    item.ProductID,
			BatchID:      item.BatchID,
			OffboardQty:  item.OffboardQty,
			StoredAt:     batch.StoredAt,
			DurationDays: durationDays,
			StorageCost:  storageCost,
			BuyingPrice:  batchProduct.BillingPrice,
			SellingPrice: item.SellingPrice,
			TotalSelling: totalSellingPrice,
			BatchStatus:  status,
		})

		totalStorage += storageCost
		totalBuying += totalBuyingPrice
		totalSelling += totalSellingPrice
	}

	var otherExpenses float64
	for _, e := range expenses {
		otherExpenses += e.Amount
	}

	billing := &models.Billing{
		ID:            primitive.NewObjectID(),
		Items:         billingItems,
		EndDate:       endDate,
		RentPerUnit:   rentPerUnit,
		TotalStorage:  totalStorage,
		TotalBuying:   totalBuying,
		TotalSelling:  totalSelling,
		OtherExpenses: otherExpenses,
		Margin:        totalSelling - (totalStorage + totalBuying + otherExpenses),
		CreatedAt:     time.Now(),
	}

	if _, err := r.col.InsertOne(ctx, billing); err != nil {
		return nil, err
	}

	return billing, nil
}

// GetByID fetches a billing record by ID
func (r *BillingRepo) GetByID(ctx context.Context, id string) (*models.Billing, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var billing models.Billing
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&billing)
	if err != nil {
		return nil, err
	}
	return &billing, nil
}

// GetAll fetches all billing records
func (r *BillingRepo) GetAll(ctx context.Context) ([]models.Billing, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []models.Billing
	for cursor.Next(ctx) {
		var b models.Billing
		if err := cursor.Decode(&b); err != nil {
			return nil, err
		}
		bills = append(bills, b)
	}
	return bills, nil
}
