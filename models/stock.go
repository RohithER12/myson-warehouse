package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductStock struct {
	ProductID     primitive.ObjectID `json:"product_id"`
	ProductName   string             `json:"product_name"`
	TotalQuantity int                `json:"total_quantity"`
	TotalRent     float64            `json:"total_rent"`
	TotalSpace    float64            `json:"total_space"`
	Currency      string             `json:"currency"`
}



//helper Struct
type BatchProduct struct {
	WarehouseID string             `bson:"warehouse_id"`
	ProductID   primitive.ObjectID `bson:"product_id"`
	Quantity    int                `bson:"quantity"`
	StoredAt    time.Time          `bson:"stored_at"`
}
