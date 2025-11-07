package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	SupplierID  uint           `gorm:"not null;index" json:"supplier_id"`
	Supplier    Supplier       `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"supplier"`
	Category    string         `gorm:"type:varchar(255)" json:"category"`
	StorageArea float64        `gorm:"type:decimal(10,2);not null" json:"storage_area"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
type ProductData struct {
	ID          uint      ` json:"id"`
	Name        string    ` json:"name"`
	SupplierID  uint      ` json:"supplier_id"`
	Supplier    Supplier  ` json:"supplier"`
	Category    string    ` json:"category"`
	StorageArea float64   ` json:"storage_area"`
	CreatedAt   time.Time ` json:"created_at"`
	UpdatedAt   time.Time ` json:"updated_at"`
}
