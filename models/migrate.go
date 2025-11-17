package models

import "time"

type ExcelProductRow struct {
    ProductName     string
    ProductCategory string
    SupplierName    string
    StorageArea     float64
    WarehouseName   string
    PurchasePrice   float64
    Quantity        int
    PurchaseDate    time.Time
}