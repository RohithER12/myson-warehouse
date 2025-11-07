package repo

import (
	"context"
	"fmt"
	"log"
	"time"
	dbconn "warehouse/config/dbConn"
	"warehouse/models"
)

type WarehouseRepo struct {
}

// NewWarehouseRepo initializes the repository
func NewWarehouseRepo() *WarehouseRepo {
	return &WarehouseRepo{}
}

// Create inserts a new warehouse
func (r *WarehouseRepo) Create(ctx context.Context, warehouse *models.Warehouse) (uint, error) {
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Warehouse")

	if err := db.Table(table).Create(&warehouse).Error; err != nil {
		return 0, fmt.Errorf("failed to create warehouse: %w", err)
	}

	log.Printf("🏠 New warehouse created: ID=%d, Name=%s", warehouse.ID, warehouse.Name)
	return warehouse.ID, nil
}

// GetByID fetches a warehouse by ID
func (r *WarehouseRepo) GetByID(ctx context.Context, id uint) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Warehouse")

	if err := db.Table(table).
		Preload("RentConfig").
		First(&warehouse, id).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch warehouse ID %d: %w", id, err)
	}
	return &warehouse, nil
}

// GetAll fetches all warehouses
func (r *WarehouseRepo) GetAll(ctx context.Context) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Warehouse")

	if err := db.Table(table).
		Preload("RentConfig").
		Find(&warehouses).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch warehouses: %w", err)
	}
	return warehouses, nil
}

// Update modifies a warehouse record
func (r *WarehouseRepo) Update(ctx context.Context, warehouse *models.Warehouse) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	warehouseTable := ns.TableName("Warehouse")
	rentRateTable := ns.TableName("RentRate")

	var existing models.Warehouse

	// Fetch existing warehouse with RentConfig
	if err := db.Table(warehouseTable).
		Preload("RentConfig").
		First(&existing, warehouse.ID).Error; err != nil {
		return fmt.Errorf("warehouse not found: %w", err)
	}

	// Update warehouse (excluding RentConfig relationship)
	if err := db.Table(warehouseTable).
		Where("id = ?", warehouse.ID).
		Omit("RentConfig").
		Updates(warehouse).Error; err != nil {
		return fmt.Errorf("failed to update warehouse: %w", err)
	}

	// ✅ Update RentConfig if exists
	if existing.RentConfig.ID != 0 {
		warehouse.RentConfig.ID = existing.RentConfig.ID
		if err := db.Table(rentRateTable).
			Where("id = ?", existing.RentConfig.ID).
			Updates(&warehouse.RentConfig).Error; err != nil {
			return fmt.Errorf("failed to update rent config: %w", err)
		}
	}

	log.Printf("✅ Warehouse updated: ID=%d", warehouse.ID)
	return nil
}

// Delete removes a warehouse
func (r *WarehouseRepo) Delete(ctx context.Context, id uint) error {
	db := dbconn.DB.WithContext(ctx)
	ns := db.NamingStrategy
	table := ns.TableName("Warehouse")

	if err := db.Table(table).Delete(&models.Warehouse{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete warehouse ID %d: %w", id, err)
	}

	log.Printf("🗑️ Warehouse deleted: ID=%d", id)
	return nil
}
