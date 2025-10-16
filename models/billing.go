package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Billing struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Items         []BillingItem      `bson:"items" json:"items"`
	EndDate       time.Time          `bson:"end_date" json:"end_date"`
	RentPerUnit   float64            `bson:"rent_per_unit" json:"rent_per_unit"`
	TotalStorage  float64            `bson:"total_storage" json:"total_storage"`
	TotalBuying   float64            `bson:"total_buying" json:"total_buying"`
	TotalSelling  float64            `bson:"total_selling" json:"total_selling"`
	OtherExpenses float64            `bson:"other_expenses" json:"other_expenses"`
	Margin        float64            `bson:"margin" json:"margin"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

type BillingItem struct {
	ProductID    string    `bson:"product_id" json:"product_id"`
	BatchID      string    `bson:"batch_id" json:"batch_id"`
	OffboardQty  int       `bson:"offboard_quantity" json:"offboard_quantity"`
	StoredAt     time.Time `bson:"stored_at" json:"stored_at"`
	DurationDays float64   `bson:"duration_days" json:"duration_days"`
	StorageCost  float64   `bson:"storage_cost" json:"storage_cost"`
	BuyingPrice  float64   `bson:"buying_price" json:"buying_price"`
	SellingPrice float64   `bson:"selling_price" json:"selling_price"`
	TotalSelling float64   `bson:"total_selling" json:"total_selling"`
	BatchStatus  string    `bson:"batch_status" json:"batch_status"`
}

type Expense struct {
	Type   string  `bson:"type" json:"type"`
	Amount float64 `bson:"amount" json:"amount"`
	Notes  string  `bson:"notes" json:"notes"`
}
