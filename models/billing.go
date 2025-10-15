package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BillingBatch struct {
	BatchID  string `bson:"batch_id" json:"batch_id"`
	Quantity int    `bson:"quantity" json:"quantity"`
}

type Billing struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID        string             `bson:"product_id" json:"product_id"`
	OffboardQuantity int                `bson:"offboard_quantity" json:"offboard_quantity"`
	StartDate        time.Time          `bson:"start_date" json:"start_date"`
	EndDate          time.Time          `bson:"end_date" json:"end_date"`
	StorageDuration  float64            `bson:"storage_duration" json:"storage_duration"`
	RentPerUnit      float64            `bson:"rent_per_unit" json:"rent_per_unit"`
	StorageCost      float64            `bson:"storage_cost" json:"storage_cost"`
	OtherExpenses    float64            `bson:"other_expenses" json:"other_expenses"`
	TotalCost        float64            `bson:"total_cost" json:"total_cost"`
	TotalSelling     float64            `bson:"total_selling" json:"total_selling"`
	Margin           float64            `bson:"margin" json:"margin"`
	BatchesUsed      []BillingBatch     `bson:"batches_used" json:"batches_used"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}
