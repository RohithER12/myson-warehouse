package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductAnalytics struct {
	ProductID      primitive.ObjectID `bson:"product_id" json:"product_id"`
	ProductName    string             `bson:"product_name" json:"product_name"`
	TotalStored    int                `bson:"total_stored" json:"total_stored"`
	TotalReleased  int                `bson:"total_released" json:"total_released"`
	AvgStorageTime float64            `bson:"avg_storage_time" json:"avg_storage_time"` // days
	TotalProfit    float64            `bson:"total_profit" json:"total_profit"`
	IsFastMoving   bool               `bson:"is_fast_moving" json:"is_fast_moving"`
}
