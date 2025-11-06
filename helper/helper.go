package helper

import (
	"time"
	"warehouse/models"
)

func GetDurationRange(duration string) (time.Time, time.Time) {
	now := time.Now()
	var start time.Time

	switch duration {
	case "lastweek":
		start = now.AddDate(0, 0, -7)
	case "lastmonth":
		start = now.AddDate(0, -1, 0)
	case "lastyear":
		start = now.AddDate(-1, 0, 0)
	default:
		start = now.AddDate(0, 0, -30)
	}

	return start, now
}

func GetOrCreateProductData(m map[uint]*models.ProductWiseData, productID uint) *models.ProductWiseData {
	if m[productID] == nil {
		m[productID] = &models.ProductWiseData{}
	}
	return m[productID]
}
