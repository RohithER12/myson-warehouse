package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Batch struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	WarehouseID string              `bson:"warehouse_id" json:"warehouse_id"`
	StoredAt    time.Time           `bson:"stored_at" json:"stored_at"`
	Status      string              `bson:"status" json:"status"` // "active", "partially_offboarded", "fully_offboarded"
	Products    []BatchProductEntry `bson:"products" json:"products"`
}

type BatchProductEntry struct {
	ProductID      primitive.ObjectID `bson:"product_id" json:"product_id"`
	BillingPrice   float64            `bson:"billing_price" json:"billing_price"`
	SellingPrice   float64            `bson:"selling_price" json:"selling_price"`
	Quantity       int                `bson:"quantity" json:"quantity"`
	LastOffboarded *time.Time         `bson:"last_offboarded,omitempty" json:"last_offboarded,omitempty"`
}



