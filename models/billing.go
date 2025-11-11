package models

import (
	"time"

	"gorm.io/gorm"
)

type Billing struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Items         []BillingItem  `gorm:"foreignKey:BillingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items"`
	TotalRent     float64        `gorm:"type:decimal(12,2)" json:"total_rent"`
	TotalStorage  float64        `gorm:"type:decimal(12,2)" json:"total_storage"`
	TotalBuying   float64        `gorm:"type:decimal(12,2)" json:"total_buying"`
	TotalSelling  float64        `gorm:"type:decimal(12,2)" json:"total_selling"`
	OtherExpenses float64        `gorm:"type:decimal(12,2)" json:"other_expenses"`
	Margin        float64        `gorm:"type:decimal(12,2)" json:"margin"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type BillingItem struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	BillingID    uint           `gorm:"not null;index" json:"billing_id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	BatchID      uint           `gorm:"not null;index" json:"batch_id"`
	OffboardQty  int            `gorm:"not null" json:"offboard_quantity"`
	DurationDays float64        `gorm:"type:decimal(10,2)" json:"duration_days"`
	StorageCost  float64        `gorm:"type:decimal(12,2)" json:"storage_cost"`
	BuyingPrice  float64        `gorm:"type:decimal(10,2)" json:"buying_price"`
	SellingPrice float64        `gorm:"type:decimal(10,2)" json:"selling_price"`
	TotalSelling float64        `gorm:"type:decimal(10,2)" json:"total_selling"`
	BatchStatus  string         `gorm:"type:varchar(50)" json:"batch_status"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type BillingItemCoreData struct {
	ID           uint        `json:"id"`
	Product      ProductCore `json:"product"`
	BatchID      uint        `json:"batch_id"`
	OffboardQty  int         ` json:"offboard_quantity"`
	DurationDays float64     `json:"duration_days"`
	StorageCost  float64     ` json:"storage_cost"`
	BuyingPrice  float64     `json:"buying_price"`
	SellingPrice float64     `json:"selling_price"`
	TotalSelling float64     `json:"total_selling"`
	BatchStatus  string      ` json:"batch_status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   ` json:"updated_at"`
}

type Expense struct {
	Type   string  `bson:"type" json:"type"`
	Amount float64 `bson:"amount" json:"amount"`
	Notes  string  `bson:"notes" json:"notes"`
}

type BillingCoreData struct {
	ID            uint      `json:"id"`
	TotalRent     float64   `json:"total_rent"`
	TotalStorage  float64   `json:"total_storage"`
	TotalBuying   float64   `json:"total_buying"`
	TotalSelling  float64   `json:"total_selling"`
	OtherExpenses float64   ` json:"other_expenses"`
	Margin        float64   ` json:"margin"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
type BillingCoreDataWithProducts struct {
	ID            uint                  `json:"id"`
	Products      []BillingItemCoreData `json:"products"`
	TotalRent     float64               `json:"total_rent"`
	TotalStorage  float64               `json:"total_storage"`
	TotalBuying   float64               `json:"total_buying"`
	TotalSelling  float64               `json:"total_selling"`
	OtherExpenses float64               ` json:"other_expenses"`
	Margin        float64               ` json:"margin"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}
type BillingItemInput struct {
	ProductID    string  `json:"product_id"`
	BatchID      string  `json:"batch_id"`
	OffboardQty  int     `json:"offboard_quantity"`
	SellingPrice float64 `json:"selling_price"`
}
type BillingInput struct {
	Items    []BillingItemInput `json:"items"`
	Expenses []Expense          `json:"expenses"`
}
