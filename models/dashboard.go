package models

import (
	"time"

	"gorm.io/gorm"
)

type Profit struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	BatchID   uint           `gorm:"not null;index" json:"batch_id"`
	Batch     Batch          `gorm:"foreignKey:BatchID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"batch"`
	ProductID uint           `gorm:"not null;index" json:"product_id"`
	Product   Product        `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	NetProfit float64        `gorm:"type:decimal(12,2);not null;default:0.00" json:"net_profit"`
	Profit    float64        `gorm:"type:decimal(12,2);not null;default:0.00" json:"profit"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
