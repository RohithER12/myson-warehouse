package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Warehouse struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Location      string             `bson:"location" json:"location"`
	TotalArea     float64            `bson:"total_area" json:"total_area"` // in square feet
	AvailableArea float64            `bson:"available_area" json:"available_area"`
	RentConfig    RentRate           `bson:"rent_config" json:"rent_config"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type RentRate struct {
	RatePerSqft  float64 `bson:"rate_per_sqft" json:"rate_per_sqft"`
	Currency     string `bson:"currency" json:"currency"`
	BillingCycle string `bson:"billing_cycle" json:"billing_cycle"`
}
