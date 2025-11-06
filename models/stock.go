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

// helper Struct
type BatchProduct struct {
	WarehouseID string             `bson:"warehouse_id"`
	ProductID   primitive.ObjectID `bson:"product_id"`
	Quantity    int                `bson:"quantity"`
	StoredAt    time.Time          `bson:"stored_at"`
}

type ProductStockView struct {
	BatchID       uint       `json:"batch_id"`
	WarehouseID   uint       `json:"warehouse_id"`
	WarehouseName string     `json:"warehouse_name"`
	ProductID     uint       `json:"product_id"`
	ProductName   string     `json:"product_name"`
	Category      string     `json:"category"`
	SupplierName  string     `json:"supplier_name"`
	StorageArea   float64    `json:"storage_area"`
	Quantity      int        `json:"quantity"`
	StockQuantity int        `json:"stock_quantity"`
	BillingPrice  float64    `json:"billing_price"`
	CreatedAt     time.Time  `json:"created_at"`
	LastUpdated   *time.Time `json:"last_updated,omitempty"`
	RatePerSqft   float64    `json:"rate_per_sqft"`
	Currency      string     `json:"currency"`
	BillingCycle  string     `json:"billing_cycle"`
}
type BasicProductStockView struct {
	WarehouseID         uint    `json:"warehouse_id"`
	WarehouseName       string  `json:"warehouse_name"`
	ProductID           uint    `json:"product_id"`
	ProductName         string  `json:"product_name"`
	Category            string  `json:"category"`
	AverageStorageArea  float64 `json:"average_storage_area"`
	StockQuantity       int     `json:"stock_quantity"`
	AverageBillingPrice float64 `json:"average_billing_price"`
	AverageRatePerSqft  float64 `json:"average_rate_per_sqft"`
	Currency            string  `json:"currency"`
	BillingCycle        string  `json:"billing_cycle"`
}
