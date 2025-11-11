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
type BatchProductCoreData struct {
	ProductID      uint        `json:"product_id"`
	Product        ProductCore `json:"product"`
	BillingPrice   float64     ` json:"billing_price"`
	Quantity       int         `json:"quantity"`
	StockQuantity  int         ` json:"stock_quantity"`
	CreatedAt      time.Time   `json:"created_at"`
	LastOffboarded *time.Time  `json:"last_offboarded,omitempty"`
	LastUpdated    *time.Time  ` json:"last_updated,omitempty"`
}

type ProductCore struct {
	ID          uint      `json:"id"`
	Name        string    ` json:"name"`
	SupplierID  uint      ` json:"supplier_id"`
	Category    string    ` json:"category"`
	StorageArea float64   `json:"storage_area"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time ` json:"updated_at"`
}
type BatchCoreData struct {
	ID               uint      `json:"id"`
	WarehouseID      uint      `json:"warehouse_id"`
	StoredAt         time.Time ` json:"stored_at"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	BatchStock       int       `json:"batch_stock"`
	AvailableStock   int       `json:"available_stock"`
	OffBoardedAmount float64   `json:"off_boarded_amount"`
	OnBoardedAmount  float64   `json:"on_boarded_amount"`
}
type BatchCoreDataWithProducts struct {
	ID               uint                   `json:"id"`
	WarehouseID      uint                   `json:"warehouse_id"`
	Product          []BatchProductCoreData ` json:"product_data"`
	StoredAt         time.Time              ` json:"stored_at"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	BatchStock       int                    `json:"batch_stock"`
	AvailableStock   int                    `json:"available_stock"`
	OffBoardedAmount float64                `json:"off_boarded_amount"`
	OnBoardedAmount  float64                `json:"on_boarded_amount"`
}
