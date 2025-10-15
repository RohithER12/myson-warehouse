package repo

import (
	"context"
	"fmt"
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

// Expense represents a single extra cost
type Expense struct {
	Type   string  `bson:"type" json:"type"` // e.g., "transport", "handling"
	Amount float64 `bson:"amount" json:"amount"`
	Notes  string  `bson:"notes" json:"notes"`
}

// GenerateBilling creates a billing record for offboarding a product quantity
func (r *BillingRepo) GenerateBilling(
	ctx context.Context,
	productID string,
	offboardQty int,
	startDate, endDate time.Time,
	rentPerUnit float64,
	expenses []Expense,
) (*models.Billing, error) {

	// 1️⃣ Fetch product to get StorageArea
	productRepo := NewProductRepo()
	product, err := productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// 2️⃣ Fetch batches for the product (FIFO)
	stockRepo := NewStockRepo()
	batches, err := stockRepo.GetBatchesByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	remaining := offboardQty
	var storageCost, totalBuying, totalSelling float64
	usedBatches := []models.BillingBatch{}

	// Compute duration once
	duration := endDate.Sub(startDate).Hours() / 24
	if duration < 1 {
		duration = 1
	}

	for _, batch := range batches {
		if remaining <= 0 {
			break
		}

		var batchProduct *models.BatchProductEntry
		for i := range batch.Products {
			if batch.Products[i].ProductID.Hex() == productID {
				batchProduct = &batch.Products[i]
				break
			}
		}
		if batchProduct == nil || batchProduct.Quantity == 0 {
			continue
		}

		deduct := batchProduct.Quantity
		if deduct > remaining {
			deduct = remaining
		}

		// Storage cost = rentPerUnit * StorageArea (from product) * quantity * duration
		storageCost += rentPerUnit * product.StorageArea * float64(deduct)

		totalBuying += batchProduct.BillingPrice * float64(deduct)
		totalSelling += batchProduct.SellingPrice * float64(deduct)

		usedBatches = append(usedBatches, models.BillingBatch{
			BatchID:  batch.ID.Hex(),
			Quantity: deduct,
		})

		// Deduct quantity from batch
		batchProduct.Quantity -= deduct
		if batchProduct.Quantity == 0 {
			batch.Status = "fully_offboarded"
		} else {
			batch.Status = "partially_offboarded"
		}

		remaining -= deduct

		// Update batch in DB
		if err := stockRepo.UpdateBatch(ctx, &batch); err != nil {
			return nil, err
		}
	}

	if remaining > 0 {
		return nil, fmt.Errorf("not enough quantity in batches for product %s", productID)
	}

	// 3️⃣ Sum other expenses
	var otherExpenses float64
	for _, e := range expenses {
		otherExpenses += e.Amount
	}

	billing := &models.Billing{
		ID:               primitive.NewObjectID(),
		ProductID:        productID,
		OffboardQuantity: offboardQty,
		StartDate:        startDate,
		EndDate:          endDate,
		StorageDuration:  duration,
		RentPerUnit:      rentPerUnit,
		StorageCost:      storageCost,
		OtherExpenses:    otherExpenses,
		TotalCost:        storageCost + totalBuying + otherExpenses,
		TotalSelling:     totalSelling,
		Margin:           totalSelling - (storageCost + totalBuying + otherExpenses),
		BatchesUsed:      usedBatches,
		CreatedAt:        time.Now(),
	}

	// Insert billing record
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

// Helper functions
func sumExpenseByType(exp []Expense, typ string) float64 {
	var total float64
	for _, e := range exp {
		if typ == "" || e.Type == typ {
			total += e.Amount
		}
	}
	return total
}

// excludes returns expenses excluding a given type
func excludes(exp []Expense, typ string) []Expense {
	var result []Expense
	for _, e := range exp {
		if e.Type != typ {
			result = append(result, e)
		}
	}
	return result
}
