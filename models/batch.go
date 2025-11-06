package models

import (
	"time"

	"gorm.io/gorm"
)

// type Batch struct {
// 	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
// 	WarehouseID string              `bson:"warehouse_id" json:"warehouse_id"`
// 	StoredAt    time.Time           `bson:"stored_at" json:"stored_at"`
// 	Status      string              `bson:"status" json:"status"` // "active", "partially_offboarded", "fully_offboarded"
// 	Products    []BatchProductEntry `bson:"products" json:"products"`
// }

// type BatchProductEntry struct {
// 	ProductID      primitive.ObjectID `bson:"product_id" json:"product_id"`
// 	BillingPrice   float64            `bson:"billing_price" json:"billing_price"`
// 	SellingPrice   float64            `bson:"selling_price" json:"selling_price"`
// 	Quantity       int                `bson:"quantity" json:"quantity"`
// 	StockQuantity  int                `bson:"stock_quantity" json:"stock_quantity"`
// 	LastOffboarded *time.Time         `bson:"last_offboarded,omitempty" json:"last_offboarded,omitempty"`
// 	LastUpdated    *time.Time         `bson:"last_updated,omitempty" json:"last_updated,omitempty"`
// }
// type BatchData struct {
// 	ID          primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
// 	WarehouseID string                  `bson:"warehouse_id" json:"warehouse_id"`
// 	StoredAt    time.Time               `bson:"stored_at" json:"stored_at"`
// 	Status      string                  `bson:"status" json:"status"` // "active", "partially_offboarded", "fully_offboarded"
// 	Products    []BatchDataProductEntry `bson:"products" json:"products"`
// }

//	type BatchDataProductEntry struct {
//		ProductID      string     `bson:"product_id" json:"product_id"`
//		BillingPrice   float64    `bson:"billing_price" json:"billing_price"`
//		SellingPrice   float64    `bson:"selling_price" json:"selling_price"`
//		Quantity       int        `bson:"quantity" json:"quantity"`
//		StockQuantity  int        `bson:"stock_quantity" json:"stock_quantity"`
//		LastOffboarded *time.Time `bson:"last_offboarded,omitempty" json:"last_offboarded,omitempty"`
//		LastUpdated    *time.Time `bson:"last_updated,omitempty" json:"last_updated,omitempty"`
//	}
type Batch struct {
	ID          uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	WarehouseID uint                `gorm:"not null" json:"warehouse_id"` // 🔧 changed from string → uint
	Warehouse   Warehouse           `gorm:"foreignKey:WarehouseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"warehouse"`
	StoredAt    time.Time           `gorm:"autoCreateTime" json:"stored_at"`
	Status      string              `gorm:"type:varchar(50)" json:"status"` // active, partially_offboarded, fully_offboarded
	Products    []BatchProductEntry `gorm:"foreignKey:BatchID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"products"`
	CreatedAt   time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `gorm:"index" json:"-"`
}

type BatchProductEntry struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID        uint       `gorm:"not null;index" json:"batch_id"`
	ProductID      uint       `gorm:"not null;index" json:"product_id"` // 🔧 changed from string → uint
	Product        Product    `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	BillingPrice   float64    `gorm:"type:decimal(10,2)" json:"billing_price"`
	SellingPrice   float64    `gorm:"type:decimal(10,2)" json:"selling_price"`
	Quantity       int        `gorm:"not null" json:"quantity"`
	StockQuantity  int        `gorm:"not null" json:"stock_quantity"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	LastOffboarded *time.Time `json:"last_offboarded,omitempty"`
	LastUpdated    *time.Time `gorm:"autoUpdateTime" json:"last_updated,omitempty"`
}
