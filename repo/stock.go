package repo

import (
	"context"
	"strings"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductStockRepo struct {
	col          *mongo.Collection
	productCol   *mongo.Collection
	warehouseCol *mongo.Collection
}

// NewStockRepo initializes the repository
func NewProductStockRepo() *ProductStockRepo {
	return &ProductStockRepo{
		col:          dbconn.GetCollection("myson_warehouse", BatchCollection),
		productCol:   dbconn.GetCollection("myson_warehouse", productCollection),
		warehouseCol: dbconn.GetCollection("myson_warehouse", warehouseCollection),
	}
}

func (r *ProductStockRepo) GetProductStockWithRent(ctx context.Context) ([]models.ProductStock, error) {
	// Step 1: Get all batches with product + warehouse info
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$products"}},
		{{Key: "$project", Value: bson.M{
			"warehouse_id": "$warehouse_id",
			"product_id":   "$products.product_id",
			"quantity":     "$products.quantity",
			"stored_at":    "$stored_at",
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []models.BatchProduct
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	// Step 2: Collect unique product and warehouse IDs
	productIDs := make(map[primitive.ObjectID]struct{})
	warehouseIDs := make(map[primitive.ObjectID]struct{})

	for _, item := range items {
		productIDs[item.ProductID] = struct{}{}
		if wid, err := primitive.ObjectIDFromHex(item.WarehouseID); err == nil {
			warehouseIDs[wid] = struct{}{}
		}
	}

	var pidList, widList []primitive.ObjectID
	for id := range productIDs {
		pidList = append(pidList, id)
	}
	for id := range warehouseIDs {
		widList = append(widList, id)
	}

	// Step 3: Fetch products and warehouses
	productCursor, err := r.productCol.Find(ctx, bson.M{"_id": bson.M{"$in": pidList}})
	if err != nil {
		return nil, err
	}
	defer productCursor.Close(ctx)

	warehouseCursor, err := r.warehouseCol.Find(ctx, bson.M{"_id": bson.M{"$in": widList}})
	if err != nil {
		return nil, err
	}
	defer warehouseCursor.Close(ctx)

	productMap := make(map[primitive.ObjectID]models.Product)
	for productCursor.Next(ctx) {
		var p models.Product
		if err := productCursor.Decode(&p); err != nil {
			return nil, err
		}
		productMap[p.ID] = p
	}

	warehouseMap := make(map[primitive.ObjectID]models.Warehouse)
	for warehouseCursor.Next(ctx) {
		var w models.Warehouse
		if err := warehouseCursor.Decode(&w); err != nil {
			return nil, err
		}
		warehouseMap[w.ID] = w
	}

	// Step 4: Aggregate total quantity, space, and rent
	type agg struct {
		Quantity int
		Space    float64
		Rent     float64
		Currency string
	}

	aggregated := make(map[primitive.ObjectID]agg)

	now := time.Now()

	for _, item := range items {
		product := productMap[item.ProductID]
		wid, err := primitive.ObjectIDFromHex(item.WarehouseID)
		if err != nil {
			continue
		}

		warehouse, exists := warehouseMap[wid]
		if !exists {
			continue
		}

		rentPerSqft := warehouse.RentConfig.RatePerSqft
		storageArea := product.StorageArea
		currency := warehouse.RentConfig.Currency
		billingCycle := strings.ToLower(warehouse.RentConfig.BillingCycle)

		// Calculate duration in days
		daysStored := now.Sub(item.StoredAt).Hours() / 24
		if daysStored < 1 {
			daysStored = 1 // minimum 1 day rent
		}

		var rent float64
		switch billingCycle {
		case "daily":
			rent = daysStored * storageArea * rentPerSqft * float64(item.Quantity)
		case "monthly":
			months := daysStored / 30
			rent = months * storageArea * rentPerSqft * float64(item.Quantity)
		default:
			// default assume monthly
			months := daysStored / 30
			rent = months * storageArea * rentPerSqft * float64(item.Quantity)
		}

		prev := aggregated[item.ProductID]
		prev.Quantity += item.Quantity
		prev.Space += storageArea * float64(item.Quantity)
		prev.Rent += rent
		prev.Currency = currency
		aggregated[item.ProductID] = prev
	}

	// Step 5: Build final result
	var result []models.ProductStock
	for pid, data := range aggregated {
		p := productMap[pid]
		result = append(result, models.ProductStock{
			ProductID:     pid,
			ProductName:   p.Name,
			TotalQuantity: data.Quantity,
			TotalSpace:    data.Space,
			TotalRent:     data.Rent,
			Currency:      data.Currency,
		})
	}

	return result, nil
}
