package handlers

// Handler: Accept Excel and send rows to service
// func ImportProductsFromExcel(c *gin.Context) {
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
// 		return
// 	}

// 	f, err := file.Open()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
// 		return
// 	}
// 	defer f.Close()

// 	// Parse Excel
// 	xlFile, err := excelize.OpenReader(f)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid excel file"})
// 		return
// 	}

// 	sheet := xlFile.GetSheetName(0)
// 	rows, err := xlFile.GetRows(sheet)
// 	if err != nil || len(rows) < 2 {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "excel must contain header + data"})
// 		return
// 	}

// 	// Convert rows → DTO
// 	var items []models.ExcelProductRow

// 	for i := 1; i < len(rows); i++ {
// 		r := rows[i]
// 		if len(r) < 8 {
// 			continue
// 		}

// 		quantity, _ := strconv.Atoi(r[6])
// 		storageArea, _ := strconv.ParseFloat(r[3], 64)
// 		purchasePrice, _ := strconv.ParseFloat(r[5], 64)
// 		purchaseDate, _ := time.Parse("2006-01-02", r[7])

// 		items = append(items, models.ExcelProductRow{
// 			ProductName:     r[0],
// 			ProductCategory: r[1],
// 			SupplierName:    r[2],
// 			StorageArea:     storageArea,
// 			WarehouseName:   r[4],
// 			PurchasePrice:   purchasePrice,
// 			Quantity:        quantity,
// 			PurchaseDate:    purchaseDate,
// 		})
// 	}

// 	// Send to service
// 	err = h.ImportService.ProcessExcelRows(c, items)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Excel imported successfully"})
// }
