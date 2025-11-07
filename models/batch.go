package models

import (
	"time"

	"gorm.io/gorm"
)


type Batch struct {
	ID          uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	WarehouseID uint                `gorm:"not null;index" json:"warehouse_id"`
	Warehouse   Warehouse           `gorm:"foreignKey:WarehouseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"warehouse"`
	StoredAt    time.Time           `gorm:"autoCreateTime" json:"stored_at"`
	Status      string              `gorm:"type:varchar(50)" json:"status"`
	Products    []BatchProductEntry `gorm:"foreignKey:BatchID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"products"`
	CreatedAt   time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt      `gorm:"index" json:"-"`
}

type BatchProductEntry struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID        uint       `gorm:"not null;index" json:"batch_id"`
	ProductID      uint       `gorm:"not null;index" json:"product_id"`
	Product        Product    `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	BillingPrice   float64    `gorm:"type:decimal(10,2);not null" json:"billing_price"`
	SellingPrice   float64    `gorm:"type:decimal(10,2)" json:"selling_price"`
	Quantity       int        `gorm:"not null" json:"quantity"`
	StockQuantity  int        `gorm:"not null" json:"stock_quantity"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	LastOffboarded *time.Time `json:"last_offboarded,omitempty"`
	LastUpdated    *time.Time `gorm:"autoUpdateTime" json:"last_updated,omitempty"`
}
