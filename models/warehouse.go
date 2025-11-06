package models

import (
	"time"

	"gorm.io/gorm"
)

// type Warehouse struct {
// 	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
// 	Name          string             `bson:"name" json:"name"`
// 	Location      string             `bson:"location" json:"location"`
// 	TotalArea     float64            `bson:"total_area" json:"total_area"` // in square feet
// 	AvailableArea float64            `bson:"available_area" json:"available_area"`
// 	RentConfig    RentRate           `bson:"rent_config" json:"rent_config"`
// 	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
// 	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
// }

// type RentRate struct {
// 	RatePerSqft  float64 `bson:"rate_per_sqft" json:"rate_per_sqft"`
// 	Currency     string `bson:"currency" json:"currency"`
// 	BillingCycle string `bson:"billing_cycle" json:"billing_cycle"`
// }

type Warehouse struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Location      string         `gorm:"type:varchar(255)" json:"location"`
	TotalArea     float64        `gorm:"type:decimal(10,2);not null" json:"total_area"`
	AvailableArea float64        `gorm:"type:decimal(10,2);not null" json:"available_area"`
	RentConfigID  uint           `gorm:"not null" json:"rent_config_id"`
	RentConfig    RentRate       `gorm:"foreignKey:RentConfigID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"rent_config"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type RentRate struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	RatePerSqft  float64        `gorm:"type:decimal(10,2);not null" json:"rate_per_sqft"`
	Currency     string         `gorm:"type:varchar(10);default:'INR'" json:"currency"`
	BillingCycle string         `gorm:"type:varchar(50);default:'monthly'" json:"billing_cycle"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}